package repository

import (
	"database/sql"
	"fmt"
	"log"

	"user-mgmt/pkg/models"

	"github.com/google/uuid"
)

func getAllUsers(db *sql.DB) ([]models.User, error) {

	users := []models.User{}

	query := `SELECT id, email, password, name, category, dob, bio, avatar FROM public."users"`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //  to be executed after the surrounding function completes

	for rows.Next() {
		var user models.User
		// Scan copies the columns in the current row into the values pointed.
		err := rows.Scan(&user.Id, &user.Email, &user.Password, &user.Name, &user.Category, &user.DOB, &user.DOBFormatted, &user.Bio, &user.Avatar)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func GetUserById(db *sql.DB, id string) (models.User, error) {

	var user models.User // zero value of User struct

	err := db.QueryRow(`SELECT id, email, password, name, category, dob, bio, avatar FROM public."users" WHERE id = $1`, id).Scan(&user.Id, &user.Email, &user.Password, &user.Name, &user.Category, &user.DOB, &user.Bio, &user.Avatar)

	if err != nil {
		return user, err
	}

	user.DOBFormatted = user.DOB.Format("2006-01-02")
	fmt.Println("Check DOB not  DOBFormatted:", user.DOB)
	fmt.Println("Check DOB try  DOBFormatted:", user.DOB.Format("2006-01-02"))

	fmt.Println("Check DOBFormatted:", user.DOBFormatted)

	return user, nil

}

func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var user models.User

	err := db.QueryRow(`SELECT id, email, password, name, category, dob, bio, avatar FROM public."users" WHERE email = $1`, email).Scan(&user.Id, &user.Email, &user.Password, &user.Name, &user.Category, &user.DOB, &user.Bio, &user.Avatar)

	return user, err
}

func CreateUser(db *sql.DB, user models.User) error {

	log.Println("Check DB:", db)

	id, err := uuid.NewUUID()

	if err != nil {
		return err
	}

	//Convert id to string and set it on the user
	user.Id = id.String()

	// The Prepare method is used to create a prepared statement.
	stmt, err := db.Prepare(`INSERT INTO public."users" (id, email, password, name, category, dob, bio, avatar) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		log.Println("Check Error:", err)

		return err
	}

	defer stmt.Close()

	// Once a statement is prepared, Exec is used to execute it with the actual values specified.
	_, err = stmt.Exec(user.Id, user.Email, user.Password, user.Name, user.Category, user.DOB, user.Bio, user.Avatar)

	if err != nil {
		return err
	}

	return nil
}

func UpdateUser(db *sql.DB, id string, user models.User) error {
	_, err := db.Exec(`UPDATE public."users" SET name = $1, category = $2, dob = $3, bio = $4 WHERE id = $5`, user.Name, user.Category, user.DOB, user.Bio, id)

	return err
}

func UpdateUserAvatar(db *sql.DB, userID, filePath string) error {
	_, err := db.Exec(`UPDATE public."users" SET avatar = $1 WHERE id = $2`, filePath, userID)
	return err
}

func DeleteUser(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM public."users" WHERE id = $1`, id)

	return err
}
