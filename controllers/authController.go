package controllers

import (
    "context"
    "filestorage-backend/config"
    "filestorage-backend/models"
    "filestorage-backend/utils"
    "github.com/gofiber/fiber/v2"
    "golang.org/x/crypto/bcrypt"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "net/http"
    "encoding/json"
)

var (
    oauthConfig = &oauth2.Config{
        RedirectURL:  config.Get("OAUTH_REDIRECT_URL"),
        ClientID:     config.Get("OAUTH_CLIENT_ID"),
        ClientSecret: config.Get("OAUTH_CLIENT_SECRET"),
        Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
        Endpoint:     google.Endpoint,
    }
)

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func Register(c *fiber.Ctx) error {
    user := new(models.User)
    if err := c.BodyParser(user); err != nil {
        return c.Status(fiber.StatusBadRequest).SendString(err.Error())
    }

    
    if user.Email == "" || user.Name == "" || user.Password == "" {
        return c.Status(fiber.StatusBadRequest).SendString("Email, Name, and Password cannot be empty")
    }

    
    collection := config.GetMongoCollection("users")
    var existingUser models.User
    err := collection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&existingUser)
    if err == nil {
        return c.Status(fiber.StatusBadRequest).SendString("User already exists")
    } else if err != mongo.ErrNoDocuments {
        return c.Status(fiber.StatusInternalServerError).SendString("Error checking for existing user")
    }

    
    hashedPassword, err := hashPassword(user.Password)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to hash password")
    }
    user.Password = hashedPassword

   
    user.FreeVersion = true

    _, err = collection.InsertOne(context.TODO(), user)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
    }

    return c.JSON(user)
}

func Login(c *fiber.Ctx) error {
    user := new(models.User)
    if err := c.BodyParser(user); err != nil {
        return c.Status(fiber.StatusBadRequest).SendString(err.Error())
    }

    storedUser := new(models.User)
    collection := config.GetMongoCollection("users")
    err := collection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&storedUser)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials")
    }

    err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials")
    }

    token, err := utils.GenerateJWT(storedUser.Email)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to generate JWT")
    }

    return c.JSON(fiber.Map{"token": token})
}

func GoogleLogin(c *fiber.Ctx) error {
    url := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
    return c.Redirect(url)
}

func GoogleCallback(c *fiber.Ctx) error {
    code := c.Query("code")
    token, err := oauthConfig.Exchange(context.Background(), code)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to exchange token")
    }

    resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to get user info")
    }
    defer resp.Body.Close()

    var userInfo map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to decode user info")
    }

    email := userInfo["email"].(string)
    user := &models.User{}
    collection := config.GetMongoCollection("users")
    err = collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(user)
    if err == mongo.ErrNoDocuments {
        user.Name = userInfo["name"].(string)
        user.Email = email
        user.DisplayName = userInfo["given_name"].(string)
        user.Picture = userInfo["picture"].(string)
        user.FreeVersion = true
        _, err = collection.InsertOne(context.TODO(), user)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).SendString("Failed to create user")
        }
    }

    jwtToken, err := utils.GenerateJWT(email)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to generate JWT")
    }

    return c.JSON(fiber.Map{"token": jwtToken})
}

func Protected(c *fiber.Ctx) error {
    user := c.Locals("user").(string)
    return c.SendString("Hello, " + user)
}
