package core_transport_request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

// decodeBody прогоняет Decode так же, как это делает прод: тело едет в настоящем
// *http.Request, а не в поддельном io.Reader. Схема — { "a": int }.
func decodeBody(t *testing.T, body string) error {
	t.Helper()

	var dest struct {
		A int `json:"a"`
	}

	r := httptest.NewRequest("POST", "/api/v1/applications", strings.NewReader(body))

	return Decode(r, &dest)
}

// asError достаёт нашу ошибку из результата Decode. Ловля идёт в value, потому
// что Decode возвращает value: перепутаешь — тест покраснеет, а не прод.
func asError(t *testing.T, err error) Error {
	t.Helper()

	var got Error
	if !errors.As(err, &got) {
		t.Fatalf("ожидался Error, получено %T: %v", err, err)
	}

	return got
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantClass string // "" — ошибки быть не должно
		wantField string
	}{
		// хорошие тела
		{name: "валидный объект", body: `{"a":1}`},
		{name: "пробелы в хвосте", body: "{\"a\":1}  \n\t "},

		// правее первого значения лежат данные
		{name: "второй объект", body: `{"a":1}{"a":2}`, wantClass: "trailing"},
		{name: "число в хвосте", body: `{"a":1} 42`, wantClass: "trailing"},
		{name: "null в хвосте", body: `{"a":1}null`, wantClass: "trailing"},
		{name: "мусор в хвосте", body: `{"a":1} xyz`, wantClass: "trailing"},

		// класса «данных нет» и «значение №1 не завершилось»
		{name: "пустое тело", body: ``, wantClass: "empty"},
		{name: "только пробелы", body: "   \n", wantClass: "empty"},
		{name: "обрезанный объект", body: `{"a":`, wantClass: "malformed"},
		{name: "висячая запятая", body: `{"a":1,}`, wantClass: "malformed"},
		{name: "не тот тип", body: `{"a":"x"}`, wantClass: "malformed"},

		// поле не из схемы
		{name: "неизвестное поле", body: `{"a":1,"b":2}`, wantClass: "unknown_field", wantField: "b"},
		{name: "неизвестное поле вторым", body: `{"a":1,"zzz":2}`, wantClass: "unknown_field", wantField: "zzz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decodeBody(t, tt.body)

			if tt.wantClass == "" {
				if err != nil {
					t.Fatalf("ожидался nil, получено %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("ожидалась ошибка класса %q, получено nil", tt.wantClass)
			}

			got := asError(t, err)
			if string(got.Class) != tt.wantClass {
				t.Fatalf("класс %q, ожидался %q (%v)", got.Class, tt.wantClass, err)
			}
			if got.Field != tt.wantField {
				t.Fatalf("поле %q, ожидалось %q", got.Field, tt.wantField)
			}
			if got.Message == "" {
				t.Fatal("пустое Message: в логе не останется человеческого текста")
			}
		})
	}
}

// TestDecodeTrailingWithoutCause кодирует самый коварный случай: `{"a":1}null`
// разбирается БЕЗ ошибки, поэтому Cause у трейлинга пустой. Если однажды
// условие пробы напишут через `err != nil`, этот тест покраснеет.
func TestDecodeTrailingWithoutCause(t *testing.T) {
	got := asError(t, decodeBody(t, `{"a":1}null`))

	if got.Class != TrailingError {
		t.Fatalf("класс %q, ожидался %q", got.Class, TrailingError)
	}
	if got.Cause != nil {
		t.Fatalf("Cause = %v, ожидался nil: `null` разбирается без ошибки", got.Cause)
	}
}

// TestDecodeCauseUnwraps проверяет Unwrap: причина обязана быть доступна
// программе, а не только человеку в логе. Типы разные нарочно — Decoder
// и Unmarshal отчитываются о похожих проблемах по-разному.
func TestDecodeCauseUnwraps(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		target func(error) bool
	}{
		{
			name: "обрезанное тело → io.ErrUnexpectedEOF",
			body: `{"a":`,
			target: func(err error) bool {
				return errors.Is(err, io.ErrUnexpectedEOF)
			},
		},
		{
			name: "реальная синтаксическая ошибка → *json.SyntaxError",
			body: `{"a":1,}`,
			target: func(err error) bool {
				var syntax *json.SyntaxError
				return errors.As(err, &syntax)
			},
		},
		{
			name: "не тот тип → *json.UnmarshalTypeError",
			body: `{"a":"x"}`,
			target: func(err error) bool {
				var typeErr *json.UnmarshalTypeError
				return errors.As(err, &typeErr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decodeBody(t, tt.body)
			if err == nil {
				t.Fatal("ожидалась ошибка, получено nil")
			}
			if !tt.target(err) {
				t.Fatalf("причина не достаётся через errors.Is/As: Unwrap порван (%v)", err)
			}
		})
	}
}

// TestDecoderUnknownFieldMessage сторожит не нашу логику, а формат сообщения
// encoding/json. Строка написана ЛИТЕРАЛОМ намеренно: если взять константу из
// decode.go, тест будет проверять её саму на себя и промолчит, когда Go сменит
// формат. Покраснел этот тест — правь разбор в classify().
func TestDecoderUnknownFieldMessage(t *testing.T) {
	const literal = `json: unknown field `

	dec := json.NewDecoder(strings.NewReader(`{"a":1,"zzz":2}`))
	dec.DisallowUnknownFields()

	var dst struct {
		A int `json:"a"`
	}
	if err := dec.Decode(&dst); err == nil {
		t.Fatal("ожидалась ошибка на неизвестном поле")
	} else if !strings.HasPrefix(err.Error(), literal) {
		t.Fatalf("формат сообщения изменился: %q не начинается с %q", err, literal)
	} else if got := strings.Trim(strings.TrimPrefix(err.Error(), literal), `"`); got != "zzz" {
		t.Fatalf("имя поля %q, ожидалось zzz", got)
	}
}
