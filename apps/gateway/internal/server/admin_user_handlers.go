package server

import "net/http"

// 后台用户路由分发只负责按路径把请求交给更细的 handler 文件。

func (a *App) handleAdminUsers(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		a.handleAdminUserList(writer, request)
	case http.MethodPost:
		a.handleAdminUserCreate(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *App) handleAdminUserByID(writer http.ResponseWriter, request *http.Request) {
	userID, action, ok := parseAdminUserPath(request.URL.Path)
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if action == "" && request.Method == http.MethodGet {
		a.handleAdminUserDetail(writer, request, userID)
		return
	}

	if action == "service-overrides" {
		switch request.Method {
		case http.MethodGet:
			a.handleAdminUserServiceOverridesList(writer, request, userID)
		case http.MethodPut:
			a.handleAdminUserServiceOverridesReplace(writer, request, userID)
		default:
			writer.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	if action == "reset-password" && request.Method == http.MethodPost {
		a.handleAdminUserPasswordReset(writer, request, userID)
		return
	}

	if action == "status" && request.Method == http.MethodPost {
		a.handleAdminUserStatusSet(writer, request, userID)
		return
	}

	if request.Method != http.MethodPatch {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	a.handleAdminUserUpdate(writer, request, userID)
}
