package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pass-wall/passwall-server/internal/app"
	"github.com/pass-wall/passwall-server/internal/common"
	"github.com/pass-wall/passwall-server/internal/encryption"
	"github.com/pass-wall/passwall-server/internal/storage"
	"github.com/pass-wall/passwall-server/model"
	"github.com/spf13/viper"
)

// FindAll ...
func FindAllNotes(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		notes := []model.Note{}

		fields := []string{"id", "created_at", "updated_at", "note"}
		argsStr, argsInt := SetArgs(r, fields)

		notes, err = s.Notes().FindAll(argsStr, argsInt)

		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		notes = app.DecryptNotes(notes)
		common.RespondWithJSON(w, http.StatusOK, notes)
	}
}

// FindByID ...
func FindNoteByID(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			common.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		note, err := s.Notes().FindByID(uint(id))
		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		passByte, _ := base64.StdEncoding.DecodeString(note.Note)
		note.Note = string(encryption.Decrypt(string(passByte[:]), viper.GetString("server.passphrase")))

		common.RespondWithJSON(w, http.StatusOK, model.ToNoteDTO(note))
	}
}

// Create ...
func CreateNote(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var noteDTO model.NoteDTO

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&noteDTO); err != nil {
			common.RespondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
			return
		}
		defer r.Body.Close()

		rawPass := noteDTO.Note
		noteDTO.Note = base64.StdEncoding.EncodeToString(encryption.Encrypt(noteDTO.Note, viper.GetString("server.passphrase")))

		createdNote, err := s.Notes().Save(model.ToNote(noteDTO))
		if err != nil {
			common.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		createdNote.Note = rawPass

		common.RespondWithJSON(w, http.StatusOK, model.ToNoteDTO(createdNote))
	}
}

// Update ...
func UpdateNote(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			common.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		var noteDTO model.NoteDTO
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&noteDTO); err != nil {
			common.RespondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
			return
		}
		defer r.Body.Close()

		note, err := s.Notes().FindByID(uint(id))
		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		rawPass := noteDTO.Note
		noteDTO.Note = base64.StdEncoding.EncodeToString(encryption.Encrypt(noteDTO.Note, viper.GetString("server.passphrase")))

		noteDTO.ID = uint(id)
		note = model.ToNote(noteDTO)
		note.ID = uint(id)

		updatedNote, err := s.Notes().Save(note)
		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		updatedNote.Note = rawPass
		common.RespondWithJSON(w, http.StatusOK, model.ToNoteDTO(updatedNote))
	}
}

// Delete ...
func DeleteNote(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			common.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		note, err := s.Notes().FindByID(uint(id))
		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		err = s.Notes().Delete(note.ID)
		if err != nil {
			common.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		response := model.Response{http.StatusOK, "Success", "Note deleted successfully!"}
		common.RespondWithJSON(w, http.StatusOK, response)
	}
}