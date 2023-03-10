package handlers

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/andrepmagalhaes/q2bank_test/utils"
	"github.com/gofiber/fiber/v2"
)

type CreateBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	CpfCnpj	string `json:"cpf_cnpj"`
	Name string `json:"name"`
	UserType string `json:"user_type"`
}

type CreateResponse struct {
	Message string `json:"message"`
}

type LoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Token string `json:"token"`
}

func validateUserQuery(email string, cpfCnpj string, db *sql.DB) error {
	rows, err := db.Query("SELECT ID FROM public.\"Users\" WHERE email = $1 OR cpf_cnpj = $2", email, cpfCnpj)

	if err != nil {
		log.Printf("Error checking if user exists: %s", err.Error())
		return fmt.Errorf("internal server error")
	}

	if rows.Next() {
		return fmt.Errorf("user with this email or cpf_cnpj already exists")
	}

	return nil
}

func insertUserQuery(body CreateBody, db *sql.DB) error {

	if body.UserType == "bank" {
		return fmt.Errorf("user type bank is not allowed")
	}

	_, err := db.Exec("INSERT INTO public.\"Users\" (email, password, cpf_cnpj, name, user_type) VALUES ($1, $2, $3, $4, $5)", body.Email, body.Password, body.CpfCnpj, body.Name, body.UserType)

	if err != nil {
		log.Printf("Error inserting user: %s", err.Error())
		return fmt.Errorf("internal server error")
	}

	return nil
}

func findUserQuery(email string, db *sql.DB) (int, string, string, error) {
	rows, err := db.Query("SELECT id, password, user_type FROM public.\"Users\" WHERE email = $1", email)

	if err != nil {
		log.Printf("Error checking if user exists: %s", err.Error())
		return -1, "", "", fmt.Errorf("internal server error")
	}

	if !rows.Next() {
		return -1, "", "", fmt.Errorf("user not found")
	}

	var id int
	var password string
	var userType string

	err = rows.Scan(&id, &password, &userType)

	if err != nil {
		log.Printf("Error scanning user: %s", err.Error())
		return -1, "", "", fmt.Errorf("internal server error")
	}

	return id, password, userType, nil
}

func Create(c *fiber.Ctx, db *sql.DB) error {
	body := CreateBody{}
	response := CreateResponse{}

	if err := c.BodyParser(&body); err != nil {
		response.Message = err.Error()
		return c.Status(400).JSON(response)
	}

	if body.Email == "" {
		response.Message = "email is required"
		return c.Status(400).JSON(response)
	}

	if body.Password == "" {
		response.Message = "password is required"
		return c.Status(400).JSON(response)
	}

	if body.CpfCnpj == "" {
		response.Message = "cpf_cnpj is required"
		return c.Status(400).JSON(response)
	}

	if body.Name == "" {
		response.Message = "name is required"
		return c.Status(400).JSON(response)
	}

	if body.UserType == "" {
		response.Message = "user_type is required"
		return c.Status(400).JSON(response)
	}

	if body.UserType != "person" && body.UserType != "store" {
		response.Message = "user_type must be either person or store"
		return c.Status(400).JSON(response)
	}

	pwValidationMessage, pwValidationValidity := utils.ValidatePassword(body.Password)
	
	if !pwValidationValidity {
		response.Message = pwValidationMessage
		return c.Status(400).JSON(response)
	}

	hashedPassword, err := utils.HashPassword(body.Password)

	if err != nil {
		log.Printf("Error hashing password: %s", err.Error())
		response.Message = "internal server error"
		return c.Status(500).JSON(response)
	}

	err = validateUserQuery(body.Email, body.CpfCnpj, db)

	if err != nil {
		response.Message = err.Error()
		return c.Status(500).JSON(response)
	}

	err = insertUserQuery(CreateBody{Email: body.Email, Password: hashedPassword, CpfCnpj: body.CpfCnpj, Name: body.Name, UserType: body.UserType}, db)

	if err != nil {
		response.Message = err.Error()
		return c.Status(500).JSON(response)
	}

	response.Message = "created"
	return c.Status(200).JSON(response)
}

func Login(c *fiber.Ctx, db *sql.DB) error {
	body := LoginBody{}
	response := LoginResponse{}

	if err := c.BodyParser(&body); err != nil {
		response.Message = err.Error()
		return c.Status(400).JSON(response)
	}

	if body.Email == "" {
		response.Message = "email is required"
		return c.Status(400).JSON(response)
	}

	if body.Password == "" {
		response.Message = "password is required"
		return c.Status(400).JSON(response)
	}

	id, password, userType, err := findUserQuery(body.Email, db)

	if err != nil {
		response.Message = "internal server error"
		return c.Status(500).JSON(response)
	}

	if !utils.CheckPasswordHash(body.Password, password) {
		response.Message = "wrong password"
		return c.Status(400).JSON(response)
	}

	response.Token, err = utils.CreateJWT(id, userType)

	if err != nil {
		log.Printf("Error creating JWT: %s", err.Error())
		response.Message = "internal server error"
		return c.Status(500).JSON(response)
	}

	response.Message = "logged in"
	return c.Status(200).JSON(response)
}