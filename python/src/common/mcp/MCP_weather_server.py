#!/usr/bin/env python3
"""
Weather MCP Server using FastMCP framework.
Provides a 'get_weather' tool for retrieving weather information.
"""
from fastmcp import FastMCP

# Create the MCP server
mcp = FastMCP("Weather Server")

# Dummy temperature data (in a real implementation, call a weather API)
DUMMY_TEMPS = {
    "nyc": {"temp": "58°F", "condition": "Partly cloudy"},
    "new york": {"temp": "58°F", "condition": "Partly cloudy"},
    "london": {"temp": "48°F", "condition": "Rainy"},
    "san francisco": {"temp": "62°F", "condition": "Sunny"},
    "beijing": {"temp": "45°F", "condition": "Hazy"},
    "tokyo": {"temp": "55°F", "condition": "Clear"},
    "paris": {"temp": "50°F", "condition": "Overcast"},
    "sydney": {"temp": "75°F", "condition": "Sunny"},
}


@mcp.tool
def get_weather(city: str) -> str:
    """
    Get the current weather for a specified city.
    
    Args:
        city: The name of the city to get weather for (e.g., "London", "NYC", "Beijing")
    
    Returns:
        A string describing the current weather conditions.
    """
    city_lower = city.lower().strip()
    
    if city_lower in DUMMY_TEMPS:
        weather = DUMMY_TEMPS[city_lower]
        return f"The current weather in {city.title()} is {weather['temp']} and {weather['condition'].lower()}."
    else:
        # Default response for unknown cities
        return f"The current weather in {city.title()} is approximately 65°F with clear skies."


if __name__ == "__main__":
    # Run using HTTP transport on port 8000
    mcp.run(transport="streamable-http", host="0.0.0.0", port=8000)
