package features_application

import (
	"errors"
	"strings"
	"testing"
)

// valid возвращает заведомо валидный манифест. Каждый тест ломает одно поле,
// чтобы было понятно, какая именно проверка сработала.
func valid() Application {
	return Application{
		APIVersion: "platform.deploy/v1",
		Kind:       "Application",
		Metadata:   Metadata{Name: "shop-api"},
		Spec: Spec{
			Tenant:   "shop",
			Image:    "ghcr.io/owner/shop-api",
			Tag:      "3f2a1c9",
			Replicas: 2,
			Port:     8080,
			Env:      map[string]string{"LOG_LEVEL": "info"},
			Ingress:  Ingress{Host: "demo.example.ts.net", Path: "/"},
		},
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*Application)
		wantField string // "" — ошибки быть не должно
	}{
		// эталон
		{name: "валидный манифест", mutate: func(*Application) {}},

		// apiVersion / kind
		{name: "apiVersion v2", mutate: func(a *Application) { a.APIVersion = "platform.deploy/v2" }, wantField: "apiVersion"},
		{name: "apiVersion пустой", mutate: func(a *Application) { a.APIVersion = "" }, wantField: "apiVersion"},
		{name: "kind другой", mutate: func(a *Application) { a.Kind = "App" }, wantField: "kind"},

		// metadata.name
		{name: "имя с заглавными", mutate: func(a *Application) { a.Metadata.Name = "Shop-API" }, wantField: "metadata.name"},
		{name: "имя из символов", mutate: func(a *Application) { a.Metadata.Name = "!!!" }, wantField: "metadata.name"},
		{name: "имя пустое", mutate: func(a *Application) { a.Metadata.Name = "" }, wantField: "metadata.name"},
		{name: "имя с подчёркиванием", mutate: func(a *Application) { a.Metadata.Name = "shop_api" }, wantField: "metadata.name"},
		{name: "имя начинается с дефиса", mutate: func(a *Application) { a.Metadata.Name = "-shop" }, wantField: "metadata.name"},
		{name: "имя кончается дефисом", mutate: func(a *Application) { a.Metadata.Name = "shop-" }, wantField: "metadata.name"},
		{name: "имя один символ", mutate: func(a *Application) { a.Metadata.Name = "a" }},
		{name: "имя 63 символа", mutate: func(a *Application) { a.Metadata.Name = strings.Repeat("a", 63) }},
		{name: "имя 64 символа", mutate: func(a *Application) { a.Metadata.Name = strings.Repeat("a", 64) }, wantField: "metadata.name"},

		// spec.tag
		{name: "tag пустой", mutate: func(a *Application) { a.Spec.Tag = "" }, wantField: "spec.tag"},
		{name: "tag latest", mutate: func(a *Application) { a.Spec.Tag = "latest" }, wantField: "spec.tag"},
		{name: "tag Latest в другом регистре", mutate: func(a *Application) { a.Spec.Tag = "Latest" }},

		// spec.image
		{name: "image пустой", mutate: func(a *Application) { a.Spec.Image = "" }, wantField: "spec.image"},
		{name: "image с тегом", mutate: func(a *Application) { a.Spec.Image = "ghcr.io/o/api:v1" }, wantField: "spec.image"},
		{name: "image с дайджестом", mutate: func(a *Application) { a.Spec.Image = "ghcr.io/o/api@sha256:abc" }, wantField: "spec.image"},
		{name: "image с портом в реестре", mutate: func(a *Application) { a.Spec.Image = "registry:5000/api" }},
		{name: "image без слэша", mutate: func(a *Application) { a.Spec.Image = "api" }},
		{name: "image без слэша, но с тегом", mutate: func(a *Application) { a.Spec.Image = "api:v1" }, wantField: "spec.image"},

		// spec.replicas
		{name: "replicas ноль", mutate: func(a *Application) { a.Spec.Replicas = 0 }},
		{name: "replicas минус один", mutate: func(a *Application) { a.Spec.Replicas = -1 }, wantField: "spec.replicas"},

		// spec.port
		{name: "port ноль", mutate: func(a *Application) { a.Spec.Port = 0 }, wantField: "spec.port"},
		{name: "port единица", mutate: func(a *Application) { a.Spec.Port = 1 }},
		{name: "port 65535", mutate: func(a *Application) { a.Spec.Port = 65535 }},
		{name: "port 65536", mutate: func(a *Application) { a.Spec.Port = 65536 }, wantField: "spec.port"},

		// tenant: пустой допустим, подставится дефолт
		{name: "tenant пустой", mutate: func(a *Application) { a.Spec.Tenant = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := valid()
			tt.mutate(&a)

			err := a.Validate()

			if tt.wantField == "" {
				if err != nil {
					t.Fatalf("ожидалось nil, получено %v", err)
				}
				return
			}

			var ve ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("ожидался ValidationError, получено %T: %v", err, err)
			}
			if ve.Field != tt.wantField {
				t.Fatalf("поле %q, ожидалось %q (сообщение: %s)", ve.Field, tt.wantField, ve.Message)
			}
		})
	}
}

func TestValidateDefaultsTenant(t *testing.T) {
	a := valid()
	a.Spec.Tenant = ""

	if err := a.Validate(); err != nil {
		t.Fatalf("ожидалось nil, получено %v", err)
	}
	if a.Spec.Tenant != "shop" {
		t.Fatalf("tenant = %q, ожидалось shop: Validate не мутирует оригинал (получатель не указатель?)", a.Spec.Tenant)
	}
}

func TestValidateKeepsExplicitTenant(t *testing.T) {
	a := valid()
	a.Spec.Tenant = "acme"

	if err := a.Validate(); err != nil {
		t.Fatalf("ожидалось nil, получено %v", err)
	}
	if a.Spec.Tenant != "acme" {
		t.Fatalf("tenant = %q, ожидалось acme: дефолт затёр явное значение", a.Spec.Tenant)
	}
}

func TestValidateNil(t *testing.T) {
	var a *Application

	err := a.Validate()
	if err == nil {
		t.Fatal("ожидалась ошибка на nil-получателе, получено nil")
	}
}

func TestValidateErrorAsValidationError(t *testing.T) {
	var a Application // всё пустое

	err := a.Validate()
	if err == nil {
		t.Fatal("ожидалась ошибка на пустом манифесте")
	}

	var ve ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("ожидался ValidationError, получено %T", err)
	}
	if ve.Field != "apiVersion" {
		t.Fatalf("поле %q, ожидалось apiVersion", ve.Field)
	}
}
