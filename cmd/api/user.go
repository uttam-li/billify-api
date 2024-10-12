package main

import (
	"billify-api/internal/store"
	"net/http"
)

type userKey string

const userCtx userKey = "user"

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtx).(*store.User)

	user, err := app.store.Users.GetByID(r.Context(), user.ID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

type UpdateUserPayload struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {

	userData := r.Context().Value(userCtx).(*store.User)

	var payload UpdateUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		ID:        userData.ID,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
	}

	if err := app.store.Users.Update(r.Context(), user); err != nil {
		if err == store.ErrNotFound {
			app.notFoundResponse(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, user); err != nil {
		app.internalServerError(w, r, err)
	}
}
