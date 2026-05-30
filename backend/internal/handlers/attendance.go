package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"

	"github.com/matthewyoungbar/swim-attendance-app/internal/models"
)

// GET /swimmers — active user list for coaches and admins
func (h *Handler) listSwimmers(w http.ResponseWriter, r *http.Request) {
	email := emailFromCtx(r)
	user, err := h.db.GetUser(r.Context(), email)
	if err != nil || user == nil || (!user.IsCoach && !user.IsAdmin) {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}
	users, err := h.db.ListUsers(r.Context())
	if err != nil {
		log.Printf("ERROR listSwimmers: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	type swimmer struct {
		Email         string `json:"email"`
		FirstName     string `json:"firstName"`
		LastName      string `json:"lastName"`
		PreferredName string `json:"preferredName,omitempty"`
	}
	result := make([]swimmer, 0, len(users))
	for _, u := range users {
		if u.IsActive {
			result = append(result, swimmer{
				Email:         u.Email,
				FirstName:     u.FirstName,
				LastName:      u.LastName,
				PreferredName: u.PreferredName,
			})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].LastName != result[j].LastName {
			return result[i].LastName < result[j].LastName
		}
		return result[i].FirstName < result[j].FirstName
	})
	jsonOK(w, result)
}

// GET /practices/all — all practices (no time filter) for coaches and admins
func (h *Handler) listAllPractices(w http.ResponseWriter, r *http.Request) {
	email := emailFromCtx(r)
	user, err := h.db.GetUser(r.Context(), email)
	if err != nil || user == nil || (!user.IsCoach && !user.IsAdmin) {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}
	practices, err := h.db.GetAllPractices(r.Context())
	if err != nil {
		log.Printf("ERROR listAllPractices: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, practices)
}

// GET /attendance/{practiceId}
func (h *Handler) getAttendance(w http.ResponseWriter, r *http.Request, practiceID string) {
	email := emailFromCtx(r)
	user, err := h.db.GetUser(r.Context(), email)
	if err != nil || user == nil || (!user.IsCoach && !user.IsAdmin) {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}
	if !user.IsAdmin {
		practice, err := h.db.GetPractice(r.Context(), practiceID)
		if err != nil || practice == nil || practice.CoachID != "USER#"+email {
			jsonError(w, "forbidden", http.StatusForbidden)
			return
		}
	}
	attendance, err := h.db.GetAttendance(r.Context(), practiceID)
	if err != nil {
		log.Printf("ERROR getAttendance: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if attendance == nil {
		jsonOK(w, models.Attendance{PracticeID: practiceID, Attendees: []models.AttendeeEntry{}})
		return
	}
	jsonOK(w, attendance)
}

// PUT /attendance/{practiceId}
func (h *Handler) saveAttendance(w http.ResponseWriter, r *http.Request, practiceID string) {
	email := emailFromCtx(r)
	user, err := h.db.GetUser(r.Context(), email)
	if err != nil || user == nil || (!user.IsCoach && !user.IsAdmin) {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}
	if !user.IsAdmin {
		practice, err := h.db.GetPractice(r.Context(), practiceID)
		if err != nil || practice == nil || practice.CoachID != "USER#"+email {
			jsonError(w, "forbidden", http.StatusForbidden)
			return
		}
	}
	var req struct {
		Attendees []models.AttendeeEntry `json:"attendees"`
		Notes     string                 `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	a := models.Attendance{
		PracticeID: practiceID,
		CoachEmail: email,
		Attendees:  req.Attendees,
		Notes:      req.Notes,
	}
	if err := h.db.SaveAttendance(r.Context(), a); err != nil {
		log.Printf("ERROR saveAttendance: %v", err)
		jsonError(w, "failed to save attendance", http.StatusInternalServerError)
		return
	}
	jsonOK(w, a)
}
