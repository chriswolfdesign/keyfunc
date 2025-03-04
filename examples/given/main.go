package main

import (
	"context"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/MicahParks/keyfunc"
)

func main() {
	// Get the JWKS URL.
	//
	// This is a sample JWKS service. Visit https://jwks-service.appspot.com/ and grab a token to test this example.
	jwksURL := "https://jwks-service.appspot.com/.well-known/jwks.json"

	// Create a context that, when cancelled, ends the JWKS background refresh goroutine.
	ctx, cancel := context.WithCancel(context.Background())

	// Create the given keys.
	hmacSecret := []byte("example secret")
	const givenKID = "givenKID"
	givenKeys := map[string]keyfunc.GivenKey{
		givenKID: keyfunc.NewGivenHMAC(hmacSecret),
	}

	// Create the keyfunc options. Use an error handler that logs. Refresh the JWKS when a JWT signed by an unknown KID
	// is found or at the specified interval. Rate limit these refreshes. Timeout the initial JWKS refresh request after
	// 10 seconds. This timeout is also used to create the initial context.Context for keyfunc.Get. Add in some given
	// keys to the JWKS.
	//
	// Do not override keys with the same key ID, `kid`, in the remote JWKS. This is the default behavior.
	//
	// For a more complex example where remote keys are overwritten by given keys, see override_test.go.
	options := keyfunc.Options{
		Ctx:              ctx,
		GivenKeys:        givenKeys,
		GivenKIDOverride: false, // Default value.
		RefreshErrorHandler: func(err error) {
			log.Printf("There was an error with the jwt.Keyfunc\nError: %s", err.Error())
		},
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  time.Minute * 5,
		RefreshTimeout:    time.Second * 10,
		RefreshUnknownKID: true,
	}

	// Create the JWKS from the resource at the given URL.
	jwks, err := keyfunc.Get(jwksURL, options)
	if err != nil {
		log.Fatalf("Failed to create JWKS from resource at the given URL.\nError: %s", err.Error())
	}

	// Create a JWT signed by the give HMAC key.
	token := jwt.New(jwt.SigningMethodHS256)
	token.Header["kid"] = givenKID
	jwtB64, err := token.SignedString(hmacSecret)
	if err != nil {
		log.Fatalf("Failed to sign a JWT with the HMAC secret.\nError: %s.", err.Error())
	}

	// Parse and validate a JWT. This one is signed by the given HMAC key.
	token, err = jwt.Parse(jwtB64, jwks.Keyfunc)
	if err != nil {
		log.Fatalf("Failed to parse the JWT signed by the given HMAC key.\nError: %s.", err.Error())
	}
	if !token.Valid {
		log.Fatalf("The token signed by the given HMAC key is not valid.")
	}
	log.Println("The token signed by the given HMAC key is valid.")

	// Parse and validate a JWT. This one is signed by a non-given key and is expired.
	jwtB64 = "eyJraWQiOiJlZThkNjI2ZCIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiJXZWlkb25nIiwiYXVkIjoiVGFzaHVhbiIsImlzcyI6Imp3a3Mtc2VydmljZS5hcHBzcG90LmNvbSIsImlhdCI6MTYzMTM2OTk1NSwianRpIjoiNDY2M2E5MTAtZWU2MC00NzcwLTgxNjktY2I3NDdiMDljZjU0In0.LwD65d5h6U_2Xco81EClMa_1WIW4xXZl8o4b7WzY_7OgPD2tNlByxvGDzP7bKYA9Gj--1mi4Q4li4CAnKJkaHRYB17baC0H5P9lKMPuA6AnChTzLafY6yf-YadA7DmakCtIl7FNcFQQL2DXmh6gS9J6TluFoCIXj83MqETbDWpL28o3XAD_05UP8VLQzH2XzyqWKi97mOuvz-GsDp9mhBYQUgN3csNXt2v2l-bUPWe19SftNej0cxddyGu06tXUtaS6K0oe0TTbaqc3hmfEiu5G0J8U6ztTUMwXkBvaknE640NPgMQJqBaey0E4u0txYgyvMvvxfwtcOrDRYqYPBnA"
	token, err = jwt.Parse(jwtB64, jwks.Keyfunc)
	if err != nil {
		log.Fatalf("Failed to parse the JWT signed by a non-given key in the remote JWKS.\nError: %s.", err.Error())
	}
	if !token.Valid {
		log.Fatalf("The token signed by a non-given key in the remote JWKS is not valid.")
	}
	log.Println("The token signed by a non-given key in the remote JWKS is valid.")

	// End the background refresh goroutine when it's no longer needed.
	cancel()

	// This will be ineffectual because the line above this canceled the parent context.Context.
	// This method call is idempotent similar to context.CancelFunc.
	jwks.EndBackground()
}
