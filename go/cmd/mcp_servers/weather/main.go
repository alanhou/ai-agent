package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Weather data structure
type Weather struct {
	Temp      string
	Condition string
}

// Dummy weather data (in a real implementation, call a weather API)
var weatherData = map[string]Weather{
	"nyc":           {Temp: "58°F", Condition: "Partly cloudy"},
	"new york":      {Temp: "58°F", Condition: "Partly cloudy"},
	"london":        {Temp: "48°F", Condition: "Rainy"},
	"san francisco": {Temp: "62°F", Condition: "Sunny"},
	"beijing":       {Temp: "45°F", Condition: "Hazy"},
	"tokyo":         {Temp: "55°F", Condition: "Clear"},
	"paris":         {Temp: "50°F", Condition: "Overcast"},
	"sydney":        {Temp: "75°F", Condition: "Sunny"},
}

// titleCase converts a string to title case using golang.org/x/text/cases
func titleCase(s string) string {
	return cases.Title(language.English).String(s)
}

func getWeather(city string) string {
	cityLower := strings.ToLower(strings.TrimSpace(city))

	if weather, ok := weatherData[cityLower]; ok {
		return fmt.Sprintf("The current weather in %s is %s and %s.",
			titleCase(city), weather.Temp, strings.ToLower(weather.Condition))
	}

	// Default response for unknown cities
	return fmt.Sprintf("The current weather in %s is approximately 65°F with clear skies.",
		titleCase(city))
}

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Weather Server",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Add the get_weather tool
	weatherTool := mcp.NewTool("get_weather",
		mcp.WithDescription("Get the current weather for a specified city."),
		mcp.WithString("city",
			mcp.Required(),
			mcp.Description("The name of the city to get weather for (e.g., 'London', 'NYC', 'Beijing')"),
		),
	)

	s.AddTool(weatherTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		city, err := request.RequireString("city")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := getWeather(city)
		return mcp.NewToolResultText(result), nil
	})

	// Start using streamable HTTP transport on port 8001 (matching Python on 8000)
	addr := ":8001"
	fmt.Printf("Starting Weather MCP Server on http://localhost%s/mcp\n", addr)
	httpServer := server.NewStreamableHTTPServer(s)
	if err := httpServer.Start(addr); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
