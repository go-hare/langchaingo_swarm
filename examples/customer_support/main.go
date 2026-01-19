package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-hare/langgraphgo_swarm/swarm"
	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Mock data structures
type Flight struct {
	DepartureAirport string
	ArrivalAirport   string
	Airline          string
	Date             string
	ID               string
}

type Hotel struct {
	Location     string
	Name         string
	Neighborhood string
	ID           string
}

type Reservation struct {
	FlightInfo Flight
	HotelInfo  Hotel
}

// Global mock data
var (
	reservations = make(map[string]*Reservation)
	flights      = []Flight{
		{
			DepartureAirport: "BOS",
			ArrivalAirport:   "JFK",
			Airline:          "Jet Blue",
			Date:             time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
			ID:               "1",
		},
	}
	hotels = []Hotel{
		{
			Location:     "New York",
			Name:         "McKittrick Hotel",
			Neighborhood: "Chelsea",
			ID:           "1",
		},
	}
)

// Flight tools
func searchFlights(departureAirport, arrivalAirport, date string) []Flight {
	// Return all flights for simplicity
	return flights
}

func bookFlight(flightID, userID string) string {
	for _, flight := range flights {
		if flight.ID == flightID {
			if reservations[userID] == nil {
				reservations[userID] = &Reservation{}
			}
			reservations[userID].FlightInfo = flight
			return "Successfully booked flight"
		}
	}
	return "Flight not found"
}

// Hotel tools
func searchHotels(location string) []Hotel {
	// Return all hotels for simplicity
	return hotels
}

func bookHotel(hotelID, userID string) string {
	for _, hotel := range hotels {
		if hotel.ID == hotelID {
			if reservations[userID] == nil {
				reservations[userID] = &Reservation{}
			}
			reservations[userID].HotelInfo = hotel
			return "Successfully booked hotel"
		}
	}
	return "Hotel not found"
}

// Create agent with tools and system prompt
func createFlightAgent(ctx context.Context, model llms.Model, transferTool swarm.HandoffToolConfig) (any, error) {
	g := graph.NewStateGraph[swarm.SwarmState]()

	g.AddNode("process", "", func(ctx context.Context, state swarm.SwarmState) (swarm.SwarmState, error) {
		// Get user ID from context (in real app, would be from config)
		userID := "user1"

		// Build system prompt with reservation info
		reservation := reservations[userID]
		systemPrompt := fmt.Sprintf(
			"You are a flight booking assistant.\n\nUser's active reservation: %+v\nToday is: %s",
			reservation,
			time.Now().Format("2006-01-02"),
		)

		messages := append([]llms.MessageContent{
			llms.TextParts("system", systemPrompt),
		}, state.Messages...)

		// Define available tools
		tools := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "search_flights",
					Description: "Search flights by departure airport, arrival airport, and date (YYYY-MM-DD)",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"departure_airport": map[string]interface{}{"type": "string"},
							"arrival_airport":   map[string]interface{}{"type": "string"},
							"date":              map[string]interface{}{"type": "string"},
						},
						"required": []string{"departure_airport", "arrival_airport", "date"},
					},
				},
			},
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "book_flight",
					Description: "Book a flight by flight ID",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"flight_id": map[string]interface{}{"type": "string"},
						},
						"required": []string{"flight_id"},
					},
				},
			},
		}

		// Add handoff tool
		handoffTool := swarm.CreateHandoffTool(transferTool)
		tools = append(tools, llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        handoffTool.Name(),
				Description: handoffTool.Description(),
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		})

		response, err := model.GenerateContent(ctx, messages, llms.WithTools(tools))
		if err != nil {
			return state, err
		}

		// Add response to messages
		aiMessage := llms.TextParts("ai", response.Choices[0].Content)
		state.Messages = append(state.Messages, aiMessage)

		return state, nil
	})

	g.SetEntryPoint("process")
	g.AddEdge("process", graph.END)

	return g.Compile()
}

func createHotelAgent(ctx context.Context, model llms.Model, transferTool swarm.HandoffToolConfig) (any, error) {
	g := graph.NewStateGraph[swarm.SwarmState]()

	g.AddNode("process", "", func(ctx context.Context, state swarm.SwarmState) (swarm.SwarmState, error) {
		userID := "user1"

		reservation := reservations[userID]
		systemPrompt := fmt.Sprintf(
			"You are a hotel booking assistant.\n\nUser's active reservation: %+v\nToday is: %s",
			reservation,
			time.Now().Format("2006-01-02"),
		)

		messages := append([]llms.MessageContent{
			llms.TextParts("system", systemPrompt),
		}, state.Messages...)

		tools := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "search_hotels",
					Description: "Search hotels by location (official city name)",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{"type": "string"},
						},
						"required": []string{"location"},
					},
				},
			},
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "book_hotel",
					Description: "Book a hotel by hotel ID",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"hotel_id": map[string]interface{}{"type": "string"},
						},
						"required": []string{"hotel_id"},
					},
				},
			},
		}

		handoffTool := swarm.CreateHandoffTool(transferTool)
		tools = append(tools, llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        handoffTool.Name(),
				Description: handoffTool.Description(),
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		})

		response, err := model.GenerateContent(ctx, messages, llms.WithTools(tools))
		if err != nil {
			return state, err
		}

		aiMessage := llms.TextParts("ai", response.Choices[0].Content)
		state.Messages = append(state.Messages, aiMessage)

		return state, nil
	})

	g.SetEntryPoint("process")
	g.AddEdge("process", graph.END)

	return g.Compile()
}

func main() {
	ctx := context.Background()

	// Initialize model
	model, err := openai.New(openai.WithModel("gpt-4"))
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Create handoff tools
	transferToHotel := swarm.HandoffToolConfig{
		AgentName:   "hotel_assistant",
		Description: "Transfer user to the hotel-booking assistant that can search for and book hotels",
	}

	transferToFlight := swarm.HandoffToolConfig{
		AgentName:   "flight_assistant",
		Description: "Transfer user to the flight-booking assistant that can search for and book flights",
	}

	// Create agents
	flightAgent, err := createFlightAgent(ctx, model, transferToHotel)
	if err != nil {
		log.Fatalf("Failed to create flight agent: %v", err)
	}

	hotelAgent, err := createHotelAgent(ctx, model, transferToFlight)
	if err != nil {
		log.Fatalf("Failed to create hotel agent: %v", err)
	}

	// Create swarm
	workflow, err := swarm.CreateSwarm(swarm.SwarmConfig{
		Agents: []swarm.Agent{
			{Name: "flight_assistant", Runnable: flightAgent, Destinations: []string{"hotel_assistant"}},
			{Name: "hotel_assistant", Runnable: hotelAgent, Destinations: []string{"flight_assistant"}},
		},
		DefaultActiveAgent: "flight_assistant",
	})
	if err != nil {
		log.Fatalf("Failed to create swarm: %v", err)
	}

	var app any
	if compiler, ok := workflow.(interface{ Compile() (any, error) }); ok {
		var compileErr error
		app, compileErr = compiler.Compile()
		if compileErr != nil {
			log.Fatalf("Failed to compile swarm: %v", compileErr)
		}
	} else {
		log.Fatal("Workflow does not support Compile()")
	}

	// Example interaction
	fmt.Println("=== Customer Support Agent Swarm ===\n")

	state := swarm.SwarmState{
		Messages: []llms.MessageContent{
			llms.TextParts("user", "I need to book a flight from Boston to New York tomorrow"),
		},
	}

	var result any
	if invoker, ok := app.(interface {
		Invoke(context.Context, swarm.SwarmState) (any, error)
	}); ok {
		var invokeErr error
		result, invokeErr = invoker.Invoke(ctx, state)
		if invokeErr != nil {
			log.Fatalf("Failed to invoke: %v", invokeErr)
		}
	}

	if resultState, ok := result.(swarm.SwarmState); ok {
		fmt.Printf("Active Agent: %s\n", resultState.ActiveAgent)
		for i, msg := range resultState.Messages {
			fmt.Printf("Message %d: %v\n", i+1, msg)
		}
	}

	fmt.Println("\n=== Customer Support Example Complete ===")
}
