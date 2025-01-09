package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"user-mgmt/pkg/models"
	"user-mgmt/pkg/repository"

	"user-mgmt/views/components"
	"user-mgmt/views/editProfile"
	"user-mgmt/views/home"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func renderComponent(w http.ResponseWriter, r *http.Request, component templ.Component) {
	err := component.Render(r.Context(), w)
	if err != nil {
		log.Println("Unable to render component:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func Homepage(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user, userId := CheckLoggedIn(w, r, store, db)
		if userId == "" {
			homepage := home.Index(&user)
			renderComponent(w, r, homepage)
			return
		}

		homepage := home.SessionedHome(&user)
		renderComponent(w, r, homepage)

		// if err := tmpl.ExecuteTemplate(w, "home.html", user); err != nil {
		// 	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		// }
	}
}

func Editpage(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Check Editpage:", store)

		user, userId := CheckLoggedIn(w, r, store, db)
		if userId == "" {
			homepage := home.Index(&user)
			renderComponent(w, r, homepage)
			return
		}

		editPage := editProfile.Index(&user)
		renderComponent(w, r, editPage)
	}
}

func UpdateProfileHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
	log.Printf("Check UpdateProfileHandler:", store)
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Check UpdateProfileHandler: 2", store)

		// Retrieve the session
		currentUserProfile, userID := CheckLoggedIn(w, r, store, db)

		log.Printf("Check currentUserProfile:", currentUserProfile)

		// Parse the form
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		var errorMessages []string

		// Collect and validate form data
		name := r.FormValue("name")
		bio := r.FormValue("bio")
		dobStr := r.FormValue("dob")

		log.Printf("Check bio:", bio)
		log.Printf("Check name:", name)
		log.Printf("Check dobStr: 2", dobStr)

		if name == "" {
			errorMessages = append(errorMessages, "Name is required.")
		}

		if dobStr == "" {
			errorMessages = append(errorMessages, "Date of birth is required.")
		}

		dob, err := time.Parse("2006-01-02", dobStr)
		if err != nil {
			errorMessages = append(errorMessages, "Invalid date format.")
		}

		// Handle validation errors
		if len(errorMessages) > 0 {
			components.AuthErrors(errorMessages).Render(r.Context(), w)

			return
		}

		// Create user struct
		user := models.User{
			Id:       userID,
			Name:     name,
			DOB:      dob,
			Bio:      bio,
			Category: currentUserProfile.Category,
		}

		// Call the repository function to update the user
		if err := repository.UpdateUser(db, userID, user); err != nil {
			errorMessages = append(errorMessages, "Failed to update user")
			components.AuthErrors(errorMessages).Render(r.Context(), w)
			log.Fatal(err)

			return
		}

		// Redirect or return success
		// Set HX-Location header and return 204 No Content status
		// w.Header().Set("HX-Location", "/")
		log.Printf("Check redirext")

		w.Header().Set("HX-Location", `/; path=/; method=GET`)
		w.WriteHeader(http.StatusNoContent)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func AvatarPage(db *sql.DB, tmpl *template.Template, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user, _ := CheckLoggedIn(w, r, store, db)

		if err := tmpl.ExecuteTemplate(w, "uploadAvatar", user); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func UploadAvatarHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		user, userID := CheckLoggedIn(w, r, store, db)

		// Initialize error messages slice
		var errorMessages []string

		// Parse the multipart form, 10 MB max upload size
		r.ParseMultipartForm(10 << 20)

		// Retrieve the file from form data
		file, handler, err := r.FormFile("avatar")
		if err != nil {
			if err == http.ErrMissingFile {
				errorMessages = append(errorMessages, "No file submitted")
			} else {
				errorMessages = append(errorMessages, "Error retrieving the file")
			}

			if len(errorMessages) > 0 {
				components.AuthErrors(errorMessages).Render(r.Context(), w)
				return
			}

		}
		defer file.Close()

		// Generate a unique filename to prevent overwriting and conflicts
		uuid, err := uuid.NewRandom()
		if err != nil {
			errorMessages = append(errorMessages, "Error generating unique identifier")
			components.AuthErrors(errorMessages).Render(r.Context(), w)

			return
		}
		filename := uuid.String() + filepath.Ext(handler.Filename) // Append the file extension

		// Create the full path for saving the file
		filePath := filepath.Join("uploads", filename)

		// Save the file to the server
		dst, err := os.Create(filePath)
		if err != nil {
			errorMessages = append(errorMessages, "Error saving the file")
			components.AuthErrors(errorMessages).Render(r.Context(), w)

			return
		}
		defer dst.Close()
		if _, err = io.Copy(dst, file); err != nil {
			errorMessages = append(errorMessages, "Error saving the file")
			components.AuthErrors(errorMessages).Render(r.Context(), w)
			return
		}

		// Update the user's avatar in the database
		//userID := r.FormValue("userID") // Assuming you pass the userID somehow
		if err := repository.UpdateUserAvatar(db, userID, filename); err != nil {
			errorMessages = append(errorMessages, "Error updating user avatar")
			components.AuthErrors(errorMessages).Render(r.Context(), w)

			log.Fatal(err)
			return
		}

		//Delete current image from the initial fetch of the user
		if user.Avatar != "" {
			oldAvatarPath := filepath.Join("uploads", user.Avatar)

			//Check if the oldPath is not the same as the new path
			if oldAvatarPath != filePath {
				if err := os.Remove(oldAvatarPath); err != nil {
					fmt.Printf("Warning: failed to delete old avatar file: %s\n", err)
				}
			}
		}

		//Navigate to the profile page after the update
		w.Header().Set("HX-Location", "/static/")
		w.WriteHeader(http.StatusNoContent)
	}
}

func LoginPage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderComponent(w, r, components.Login())
	}
}

func RegisterPage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderComponent(w, r, components.Register())
	}
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		var errorMessages []string

		// Parse the form data
		r.ParseForm()

		user.Name = r.FormValue("name")
		user.Email = r.FormValue("email")
		user.Password = r.FormValue("password")
		user.Category, _ = strconv.Atoi(r.FormValue("category"))

		// Basic validation
		if user.Name == "" {
			errorMessages = append(errorMessages, "Name is required.")
		}
		if user.Email == "" {
			errorMessages = append(errorMessages, "Email is required.")
		}
		if user.Password == "" {
			errorMessages = append(errorMessages, "Password is required.")
		}

		if len(errorMessages) > 0 {
			err := components.AuthErrors(errorMessages).Render(r.Context(), w)
			if err != nil {
				log.Println("Error rendering error component:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		log.Println("Check hashedPassword:", user.Password)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		log.Println("Check hashedPassword:", hashedPassword)

		if err != nil {
			errorMessages = append(errorMessages, "Failed to hash password.")
			components.AuthErrors(errorMessages).Render(r.Context(), w)
			return
		}
		user.Password = string(hashedPassword)
		log.Println("Check user.Password:", user.Password)

		//Set default values
		user.DOB = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		user.Bio = "Bio goes here"
		user.Avatar = ""
		// Create user in the database
		err = repository.CreateUser(db, user)

		if err != nil {
			errorMessages = append(errorMessages, "Failed to create user honest: "+err.Error())
			components.AuthErrors(errorMessages).Render(r.Context(), w)
			return
		}

		// Instead of redirecting, set HTTP status code to 204 (not content) and set 'HX-Location' header
		w.Header().Set("HX-Location", "/login")
		w.WriteHeader(http.StatusNoContent)
	}
}

func LoginHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		email := r.FormValue("email")
		password := r.FormValue("password")

		var errorMessages []string

		// Basic validation
		if email == "" {
			errorMessages = append(errorMessages, "Email is required.")
		}
		if password == "" {
			errorMessages = append(errorMessages, "Password is required.")
		}

		if len(errorMessages) > 0 {
			components.AuthErrors(errorMessages).Render(r.Context(), w)
			return
		}

		// Retrieve user by email
		user, err := repository.GetUserByEmail(db, email)
		log.Println("Retrieve user by email:", err)

		if err != nil {
			if err == sql.ErrNoRows {
				errorMessages = append(errorMessages, "Invalid email or password")
				components.AuthErrors(errorMessages).Render(r.Context(), w)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Compare the hashed password from the DB with the provided password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		log.Println("CompareHashAndPassword:", user.Password)
		log.Println("CompareHashAndPassword 2:", []byte(password))

		if err != nil {
			errorMessages = append(errorMessages, "Invalid email or password")
			components.AuthErrors(errorMessages).Render(r.Context(), w)

			return
		}

		// Create session and authenticate the user
		session, err := store.Get(r, "logged-in-user")
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		session.Values["user_id"] = user.Id
		if err := session.Save(r, w); err != nil {
			http.Error(w, "Error saving session", http.StatusInternalServerError)
			return
		}

		// Set HX-Location header and return 204 No Content status
		w.Header().Set("HX-Location", "/static/")
		w.WriteHeader(http.StatusNoContent)
	}
}

func LogoutHandler(store *sessions.CookieStore) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "logged-in-user")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Remove the user from the session
		delete(session.Values, "user_id")

		// Save the changes to the session
		if err = session.Save(r, w); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Clear the session cookie
		session.Options.MaxAge = -1
		session.Save(r, w)

		// Redirect to login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func CheckLoggedIn(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, db *sql.DB) (models.User, string) {

	session, err := store.Get(r, "logged-in-user")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return models.User{}, ""
	}

	// Check if the user_id is present in the session
	userID, ok := session.Values["user_id"]
	if !ok {
		//w.Header().Set("HX-Location", "/login")
		fmt.Println("Redirecting to /login")
		// http.Redirect(w, r, "/login", http.StatusSeeOther) // 303 required for the redirect to happen
		/* w.Header().Set("HX-Location", "/login")
		w.WriteHeader(http.StatusNoContent) */
		return models.User{}, ""
	}

	// Fetch user details from the database
	user, err := repository.GetUserById(db, userID.(string)) // Ensure that user ID handling is appropriate for your ID data type
	if err != nil {
		if err == sql.ErrNoRows {
			// No user found, possibly handle by clearing the session or redirecting to login
			session.Options.MaxAge = -1 // Clear the session
			session.Save(r, w)
			//w.Header().Set("HX-Location", "/login")
			fmt.Println("Redirecting to /login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			/* w.Header().Set("HX-Location", "/login")
			w.WriteHeader(http.StatusNoContent) */
			return models.User{}, ""
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return models.User{}, ""
	}

	return user, userID.(string)
}
