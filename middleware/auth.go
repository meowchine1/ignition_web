package middleware

import (
	"context"
	"flasher/db"
	"flasher/services"
	"net/http"
	"strings"
)

type ctxKey string

const claimsKey ctxKey = "claims"

// authenticate извлекает и проверяет JWT (cookie "access_token" либо заголовок
// Authorization: Bearer <token>) и возвращает claims. Возвращает nil, если
// токен отсутствует или невалиден. Общая логика для API- и page-middleware.
func authenticate(authService *services.AuthService, r *http.Request) *services.Claims {
	token := extractToken(r)
	if token == "" {
		return nil
	}

	claims, err := authService.ParseToken(token)
	if err != nil {
		return nil
	}

	return claims
}

// RequireAuth проверяет JWT из cookie "access_token" либо из заголовка
// Authorization: Bearer <token> (для нерб/CLI-клиентов) и кладёт claims в контекст.
func RequireAuth(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := authenticate(authService, r)
			if claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdminPage защищает HTML-страницу /admin.
// Неавторизованных пользователей перенаправляет на страницу логина, а
// авторизованных, но не обладающих ролью admin, также не пускает (redirect на /login).
// Использует существующий механизм авторизации (JWT + роль из claims).
func RequireAdminPage(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := authenticate(authService, r)
			if claims == nil || claims.Role != db.RoleAdmin {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole — использовать ПОСЛЕ RequireAuth в цепочке.
func RequireRole(role db.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if claims.Role != role {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ClaimsFromContext извлекает claims из контекста запроса.
func ClaimsFromContext(r *http.Request) (*services.Claims, bool) {
	claims, ok := r.Context().Value(claimsKey).(*services.Claims)
	return claims, ok
}

func extractToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if c, err := r.Cookie("access_token"); err == nil {
		return c.Value
	}
	return ""
}

// Chain — маленький хелпер для последовательного применения middleware.
func Chain(h http.HandlerFunc, mws ...func(http.Handler) http.Handler) http.HandlerFunc {
	var handler http.Handler = h
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler.ServeHTTP
}