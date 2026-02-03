package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/osticket-cli-go/internal/api"
	"github.com/osticket-cli-go/internal/config"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	cyan       = color.New(color.FgCyan).SprintFunc()
	green      = color.New(color.FgGreen).SprintFunc()
	yellow     = color.New(color.FgYellow).SprintFunc()
	red        = color.New(color.FgRed).SprintFunc()
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "osticket",
		Short:   "CLI tool for interacting with osTicket",
		Version: "1.0.0",
	}

	// Add commands
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(ticketCmd())
	rootCmd.AddCommand(userCmd())
	rootCmd.AddCommand(infoCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getClient() *api.Client {
	if !config.IsConfigured() {
		fmt.Fprintln(os.Stderr, red("CLI not configured. Run: osticket config set --url <url> --key <apiKey>"))
		os.Exit(1)
	}
	return api.NewClient(config.GetBaseURL(), config.GetAPIKey())
}

// ==================== CONFIG COMMANDS ====================

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

	// config set
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set configuration values",
		Run: func(cmd *cobra.Command, args []string) {
			url, _ := cmd.Flags().GetString("url")
			key, _ := cmd.Flags().GetString("key")

			if url != "" {
				if err := config.SetBaseURL(url); err != nil {
					fmt.Fprintln(os.Stderr, red("Error setting URL:"), err)
					os.Exit(1)
				}
				fmt.Println(green("✓ Base URL set"))
			}
			if key != "" {
				if err := config.SetAPIKey(key); err != nil {
					fmt.Fprintln(os.Stderr, red("Error setting API key:"), err)
					os.Exit(1)
				}
				fmt.Println(green("✓ API key set"))
			}
			if url == "" && key == "" {
				fmt.Println(yellow("Please provide --url and/or --key"))
			}
		},
	}
	setCmd.Flags().String("url", "", "osTicket API base URL")
	setCmd.Flags().String("key", "", "osTicket API key")
	cmd.AddCommand(setCmd)

	// config show
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("\n" + cyan("Configuration:"))
			url := config.GetBaseURL()
			key := config.GetAPIKey()
			urlSource, keySource := config.GetConfigSource()

			urlDisplay := url
			if url == "" {
				urlDisplay = "(not set)"
			}
			keyDisplay := key
			if key == "" {
				keyDisplay = "(not set)"
			} else if len(key) > 12 {
				keyDisplay = key[:8] + "..." + key[len(key)-4:]
			}
			fmt.Printf("  Base URL: %s [%s]\n", urlDisplay, urlSource)
			fmt.Printf("  API Key:  %s [%s]\n", keyDisplay, keySource)
			fmt.Printf("  Config file: %s\n", config.GetConfigPath())
			fmt.Printf("\n  Environment variables:\n")
			fmt.Printf("    %s\n", config.EnvBaseURL)
			fmt.Printf("    %s\n\n", config.EnvAPIKey)
		},
	}
	cmd.AddCommand(showCmd)

	// config clear
	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.Clear(); err != nil {
				fmt.Fprintln(os.Stderr, red("Error clearing config:"), err)
				os.Exit(1)
			}
			fmt.Println(green("✓ Configuration cleared"))
		},
	}
	cmd.AddCommand(clearCmd)

	return cmd
}

// ==================== TICKET COMMANDS ====================

func ticketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ticket",
		Short: "Manage tickets",
	}

	// ticket get
	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a ticket by ID or ticket number",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := getClient()
			rawOut, _ := cmd.Flags().GetBool("raw")

			// Raw output - return exact API response
			if rawOut {
				raw, err := client.GetTicketRaw(args[0])
				if err != nil {
					fmt.Fprintln(os.Stderr, red("Error:"), err)
					os.Exit(1)
				}
				fmt.Println(string(raw))
				return
			}

			// JSON output (parsed and formatted)
			data, err := client.GetTicket(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, red("Error:"), err)
				os.Exit(1)
			}

			printJSON(data)
		},
	}
	getCmd.Flags().Bool("raw", false, "Output raw API response")
	cmd.AddCommand(getCmd)

	// ticket search
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search tickets",
		Run: func(cmd *cobra.Command, args []string) {
			client := getClient()
			rawOut, _ := cmd.Flags().GetBool("raw")
			number, _ := cmd.Flags().GetString("number")
			email, _ := cmd.Flags().GetString("email")
			phone, _ := cmd.Flags().GetString("phone")
			status, _ := cmd.Flags().GetInt("status")
			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")

			// Handle search by number
			if number != "" {
				if rawOut {
					raw, err := client.GetTicketRaw(number)
					if err != nil {
						fmt.Fprintln(os.Stderr, red("Error:"), err)
						os.Exit(1)
					}
					fmt.Println(string(raw))
					return
				}
				data, err := client.GetTicket(number)
				if err != nil {
					fmt.Fprintln(os.Stderr, red("Error:"), err)
					os.Exit(1)
				}
				printJSON(data)
				return
			}

			// Handle search by email
			if email != "" {
				if rawOut {
					// Raw mode: show user lookup then tickets lookup
					raw, err := client.GetUserByEmailRaw(email)
					if err != nil {
						fmt.Fprintln(os.Stderr, red("Error getting user:"), err)
						os.Exit(1)
					}
					fmt.Println("=== User Response ===")
					fmt.Println(string(raw))
					
					raw2, err := client.GetTicketsByStatusRaw(0)
					if err != nil {
						fmt.Fprintln(os.Stderr, red("Error getting tickets:"), err)
						os.Exit(1)
					}
					fmt.Println("\n=== Tickets Response ===")
					fmt.Println(string(raw2))
					return
				}
				
				data, user, err := client.SearchTicketsByEmail(email)
				if err != nil {
					fmt.Fprintln(os.Stderr, red("Error:"), err)
					os.Exit(1)
				}
				// Include user info in response
				response := map[string]interface{}{
					"total":   data.Total,
					"tickets": data.Tickets,
				}
				if user != nil {
					response["user"] = map[string]interface{}{
						"user_id": user.UserID,
						"name":    user.Name,
						"created": user.Created,
					}
				}
				printJSON(response)
				return
			}

			if phone != "" {
				fmt.Println(yellow("Phone search requires user lookup. Please search by email or ticket number instead."))
				return
			}

			// Handle search by status or date range
			if rawOut {
				var raw []byte
				var err error
				if from != "" && to != "" {
					raw, err = client.GetTicketsByDateRangeRaw(from, to)
				} else {
					raw, err = client.GetTicketsByStatusRaw(status)
				}
				if err != nil {
					fmt.Fprintln(os.Stderr, red("Error:"), err)
					os.Exit(1)
				}
				fmt.Println(string(raw))
				return
			}

			var data *api.SimpleTicketResponse
			var err error

			if from != "" && to != "" {
				data, err = client.GetTicketsByDateRange(from, to)
			} else {
				data, err = client.GetTicketsByStatus(status)
			}

			if err != nil {
				fmt.Fprintln(os.Stderr, red("Error:"), err)
				os.Exit(1)
			}

			printJSON(data)
		},
	}
	searchCmd.Flags().Bool("raw", false, "Output raw API response")
	searchCmd.Flags().String("number", "", "Search by ticket number")
	searchCmd.Flags().String("email", "", "Search by user email")
	searchCmd.Flags().String("phone", "", "Search by user phone number")
	searchCmd.Flags().Int("status", 0, "Filter by status (0=all, 1=open, 2=resolved, 3=closed)")
	searchCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	searchCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")
	cmd.AddCommand(searchCmd)

	// ticket create
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new ticket",
		Run: func(cmd *cobra.Command, args []string) {
			client := getClient()
			jsonOut, _ := cmd.Flags().GetBool("json")

			title, _ := cmd.Flags().GetString("title")
			subject, _ := cmd.Flags().GetString("subject")
			userID, _ := cmd.Flags().GetInt("user-id")
			priority, _ := cmd.Flags().GetInt("priority")
			status, _ := cmd.Flags().GetInt("status")
			dept, _ := cmd.Flags().GetInt("dept")
			sla, _ := cmd.Flags().GetInt("sla")
			topic, _ := cmd.Flags().GetInt("topic")

			ticketID, err := client.CreateTicket(api.CreateTicketParams{
				Title:      title,
				Subject:    subject,
				UserID:     userID,
				PriorityID: priority,
				StatusID:   status,
				DeptID:     dept,
				SLAID:      sla,
				TopicID:    topic,
			})

			if err != nil {
				fmt.Fprintln(os.Stderr, red("Error:"), err)
				os.Exit(1)
			}

			if jsonOut {
				printJSON(map[string]int{"ticket_id": ticketID})
				return
			}

			fmt.Println(green("\n✓ Ticket created successfully!"))
			fmt.Printf("  Ticket ID: %d\n", ticketID)
		},
	}
	createCmd.Flags().String("title", "", "Ticket title")
	createCmd.Flags().String("subject", "", "Ticket subject/body")
	createCmd.Flags().Int("user-id", 0, "User ID")
	createCmd.Flags().Int("priority", 2, "Priority ID (1=low, 2=normal, 3=high, 4=emergency)")
	createCmd.Flags().Int("status", 1, "Status ID (1=open)")
	createCmd.Flags().Int("dept", 1, "Department ID")
	createCmd.Flags().Int("sla", 1, "SLA ID")
	createCmd.Flags().Int("topic", 1, "Topic ID")
	createCmd.Flags().Bool("json", false, "Output as JSON")
	createCmd.MarkFlagRequired("title")
	createCmd.MarkFlagRequired("subject")
	createCmd.MarkFlagRequired("user-id")
	cmd.AddCommand(createCmd)

	// ticket reply
	replyCmd := &cobra.Command{
		Use:   "reply <ticketId>",
		Short: "Reply to a ticket",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := getClient()
			jsonOut, _ := cmd.Flags().GetBool("json")

			ticketID, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, red("Invalid ticket ID"))
				os.Exit(1)
			}

			body, _ := cmd.Flags().GetString("body")
			staffID, _ := cmd.Flags().GetInt("staff-id")

			err = client.ReplyToTicket(ticketID, body, staffID)
			if err != nil {
				fmt.Fprintln(os.Stderr, red("Error:"), err)
				os.Exit(1)
			}

			if jsonOut {
				printJSON(map[string]string{"status": "success"})
				return
			}

			fmt.Println(green("\n✓ Reply sent successfully!"))
		},
	}
	replyCmd.Flags().String("body", "", "Reply body")
	replyCmd.Flags().Int("staff-id", 0, "Staff ID")
	replyCmd.Flags().Bool("json", false, "Output as JSON")
	replyCmd.MarkFlagRequired("body")
	replyCmd.MarkFlagRequired("staff-id")
	cmd.AddCommand(replyCmd)

	// ticket close
	closeCmd := &cobra.Command{
		Use:   "close <ticketId>",
		Short: "Close a ticket",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := getClient()
			jsonOut, _ := cmd.Flags().GetBool("json")

			ticketID, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, red("Invalid ticket ID"))
				os.Exit(1)
			}

			body, _ := cmd.Flags().GetString("body")
			staffID, _ := cmd.Flags().GetInt("staff-id")
			username, _ := cmd.Flags().GetString("username")
			status, _ := cmd.Flags().GetInt("status")
			team, _ := cmd.Flags().GetInt("team")
			dept, _ := cmd.Flags().GetInt("dept")
			topic, _ := cmd.Flags().GetInt("topic")

			err = client.CloseTicket(api.CloseTicketParams{
				TicketID: ticketID,
				Body:     body,
				StaffID:  staffID,
				StatusID: status,
				TeamID:   team,
				DeptID:   dept,
				TopicID:  topic,
				Username: username,
			})

			if err != nil {
				fmt.Fprintln(os.Stderr, red("Error:"), err)
				os.Exit(1)
			}

			if jsonOut {
				printJSON(map[string]string{"status": "success"})
				return
			}

			fmt.Println(green("\n✓ Ticket closed successfully!"))
		},
	}
	closeCmd.Flags().String("body", "", "Closing message")
	closeCmd.Flags().Int("staff-id", 0, "Staff ID")
	closeCmd.Flags().String("username", "", "Username")
	closeCmd.Flags().Int("status", 3, "Status ID (default: 3 for closed)")
	closeCmd.Flags().Int("team", 0, "Team ID")
	closeCmd.Flags().Int("dept", 1, "Department ID")
	closeCmd.Flags().Int("topic", 1, "Topic ID")
	closeCmd.Flags().Bool("json", false, "Output as JSON")
	closeCmd.MarkFlagRequired("body")
	closeCmd.MarkFlagRequired("staff-id")
	closeCmd.MarkFlagRequired("username")
	cmd.AddCommand(closeCmd)

	return cmd
}

// ==================== USER COMMANDS ====================

func userCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}

	// user get
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a user",
		Run: func(cmd *cobra.Command, args []string) {
			client := getClient()
			jsonOut, _ := cmd.Flags().GetBool("json")
			id, _ := cmd.Flags().GetString("id")
			email, _ := cmd.Flags().GetString("email")

			var data *api.UserData
			var err error

			if id != "" {
				data, err = client.GetUserByID(id)
			} else if email != "" {
				data, err = client.GetUserByEmail(email)
			} else {
				fmt.Fprintln(os.Stderr, red("Please provide --id or --email"))
				os.Exit(1)
			}

			if err != nil {
				fmt.Fprintln(os.Stderr, red("Error:"), err)
				os.Exit(1)
			}

			if jsonOut {
				printJSON(data)
				return
			}

			if len(data.Users) == 0 {
				fmt.Println(yellow("No user found"))
				return
			}

			displayUsers(data.Users)
		},
	}
	getCmd.Flags().String("id", "", "User ID")
	getCmd.Flags().String("email", "", "User email")
	getCmd.Flags().Bool("json", false, "Output as JSON")
	cmd.AddCommand(getCmd)

	// user create
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Run: func(cmd *cobra.Command, args []string) {
			client := getClient()
			jsonOut, _ := cmd.Flags().GetBool("json")

			name, _ := cmd.Flags().GetString("name")
			email, _ := cmd.Flags().GetString("email")
			password, _ := cmd.Flags().GetString("password")
			phone, _ := cmd.Flags().GetString("phone")
			timezone, _ := cmd.Flags().GetString("timezone")
			orgID, _ := cmd.Flags().GetInt("org-id")

			userID, err := client.CreateUser(api.CreateUserParams{
				Name:     name,
				Email:    email,
				Password: password,
				Phone:    phone,
				Timezone: timezone,
				OrgID:    orgID,
				Status:   1,
			})

			if err != nil {
				fmt.Fprintln(os.Stderr, red("Error:"), err)
				os.Exit(1)
			}

			if jsonOut {
				printJSON(map[string]int{"user_id": userID})
				return
			}

			fmt.Println(green("\n✓ User created successfully!"))
			fmt.Printf("  User ID: %d\n", userID)
		},
	}
	createCmd.Flags().String("name", "", "User name")
	createCmd.Flags().String("email", "", "User email")
	createCmd.Flags().String("password", "", "User password")
	createCmd.Flags().String("phone", "", "User phone number")
	createCmd.Flags().String("timezone", "America/New_York", "Timezone")
	createCmd.Flags().Int("org-id", 0, "Organization ID")
	createCmd.Flags().Bool("json", false, "Output as JSON")
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("email")
	createCmd.MarkFlagRequired("password")
	createCmd.MarkFlagRequired("phone")
	cmd.AddCommand(createCmd)

	return cmd
}

// ==================== INFO COMMANDS ====================

func infoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get system information",
	}

	// info departments
	deptCmd := &cobra.Command{
		Use:   "departments",
		Short: "List all departments",
		Run: func(cmd *cobra.Command, args []string) {
			client := getClient()
			jsonOut, _ := cmd.Flags().GetBool("json")

			data, err := client.GetDepartments()
			if err != nil {
				fmt.Fprintln(os.Stderr, red("Error:"), err)
				os.Exit(1)
			}

			if jsonOut {
				printJSON(data)
				return
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Name"})
			table.SetHeaderColor(
				tablewriter.Colors{tablewriter.FgCyanColor},
				tablewriter.Colors{tablewriter.FgCyanColor},
			)

			for _, dept := range data.Departments {
				table.Append([]string{strconv.Itoa(dept.ID), dept.Name})
			}

			table.Render()
		},
	}
	deptCmd.Flags().Bool("json", false, "Output as JSON")
	cmd.AddCommand(deptCmd)

	// info topics
	topicsCmd := &cobra.Command{
		Use:   "topics",
		Short: "List all help topics",
		Run: func(cmd *cobra.Command, args []string) {
			client := getClient()
			jsonOut, _ := cmd.Flags().GetBool("json")

			data, err := client.GetTopics()
			if err != nil {
				fmt.Fprintln(os.Stderr, red("Error:"), err)
				os.Exit(1)
			}

			if jsonOut {
				printJSON(data)
				return
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Topic"})
			table.SetHeaderColor(
				tablewriter.Colors{tablewriter.FgCyanColor},
				tablewriter.Colors{tablewriter.FgCyanColor},
			)

			for _, topic := range data.Topics {
				table.Append([]string{strconv.Itoa(topic.TopicID), topic.Topic})
			}

			table.Render()
		},
	}
	topicsCmd.Flags().Bool("json", false, "Output as JSON")
	cmd.AddCommand(topicsCmd)

	// info sla
	slaCmd := &cobra.Command{
		Use:   "sla",
		Short: "List all SLA plans",
		Run: func(cmd *cobra.Command, args []string) {
			client := getClient()
			jsonOut, _ := cmd.Flags().GetBool("json")

			data, err := client.GetSLAs()
			if err != nil {
				fmt.Fprintln(os.Stderr, red("Error:"), err)
				os.Exit(1)
			}

			if jsonOut {
				printJSON(data)
				return
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Name", "Grace Period"})
			table.SetHeaderColor(
				tablewriter.Colors{tablewriter.FgCyanColor},
				tablewriter.Colors{tablewriter.FgCyanColor},
				tablewriter.Colors{tablewriter.FgCyanColor},
			)

			for _, sla := range data.SLA {
				table.Append([]string{
					strconv.Itoa(sla.ID),
					sla.Name,
					strconv.Itoa(sla.GracePeriod),
				})
			}

			table.Render()
		},
	}
	slaCmd.Flags().Bool("json", false, "Output as JSON")
	cmd.AddCommand(slaCmd)

	return cmd
}

// ==================== HELPER FUNCTIONS ====================

func printJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func displayTickets(tickets [][]api.Ticket) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Number", "Subject", "Status", "Created", "User ID"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
	)
	table.SetColWidth(40)

	statusMap := map[int]string{
		1: "Open",
		2: "Resolved",
		3: "Closed",
		4: "Archived",
		5: "Deleted",
	}

	for _, ticketGroup := range tickets {
		if len(ticketGroup) == 0 {
			continue
		}
		t := ticketGroup[0]

		subject := t.Subject
		if len(subject) > 37 {
			subject = subject[:37] + "..."
		}

		status := statusMap[t.StatusID]
		if status == "" {
			status = strconv.Itoa(t.StatusID)
		}

		number := t.Number
		if number == "" {
			number = strconv.Itoa(t.TicketID)
		}

		table.Append([]string{
			number,
			subject,
			status,
			t.Created,
			strconv.Itoa(t.UserID),
		})
	}

	table.Render()
	fmt.Printf("\nTotal: %d ticket(s)\n", len(tickets))
}

func displayUsers(users []api.User) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Created"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
	)

	for _, user := range users {
		table.Append([]string{
			strconv.Itoa(user.UserID),
			user.Name,
			user.Created,
		})
	}

	table.Render()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
