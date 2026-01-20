package main

// import (
// 	"Backend/routers"
// 	"Backend/utils"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// )

// // Standalone binary to serve only the deposit email endpoint locally.
// func main() {
// 	if _, err := utils.LoadConfig(); err != nil {
// 		log.Fatalf("Failed to load config: %v", err)
// 	}

// 	r := routers.SetupEmailOnlyRouter()

// 	port := os.Getenv("EMAIL_LOCAL_PORT")
// 	if port == "" {
// 		port = "8090"
// 	}
// 	addr := ":" + port

// 	fmt.Printf("Email-only server running at http://localhost%s (POST /api/email/deposit)\n", addr)
// 	if err := http.ListenAndServe(addr, r); err != nil {
// 		log.Fatalf("Email-only server failed to start: %v", err)
// 	}
// }