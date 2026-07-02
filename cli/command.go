package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/spf13/cobra"
	foursquare "github.com/way-platform/foursquare-go"
)

// NewCommand builds the Cobra command tree for the Foursquare CLI.
func NewCommand(opts ...Option) *cobra.Command {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	root := &cobra.Command{
		Use:   "foursquare",
		Short: "Foursquare Places API CLI",
		Long:  "A command-line interface for the Foursquare Place Search and Place Details APIs.",
	}

	root.AddGroup(
		&cobra.Group{ID: "api", Title: "API Commands:"},
		&cobra.Group{ID: "auth", Title: "Authentication:"},
		&cobra.Group{ID: "utils", Title: "Utilities:"},
	)
	root.SetHelpCommandGroupID("utils")
	root.SetCompletionCommandGroupID("utils")

	root.AddCommand(
		newSearchCommand(cfg),
		newGetPlaceCommand(cfg),
		newAuthCommand(cfg),
	)
	return root
}

// newSearchCommand creates the search command.
func newSearchCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Search for nearby places",
		GroupID: "api",
		Long: `Search for places near a coordinate using the Foursquare Places API.

Results are sorted by distance ascending. The --radius is applied as a native
Foursquare radius filter.

Category IDs can be specified as comma-separated values or with multiple --category flags.
Well-known category IDs:
  EV Charging Station: 5032872391d4c4b30a586d64
  Fuel Station:        4bf58dd8d48988d113951735`,
	}
	keyFlag := cmd.Flags().String("key", "", "Foursquare API key (overrides stored credentials)")
	latFlag := cmd.Flags().Float64("lat", 0, "latitude of the search center")
	lonFlag := cmd.Flags().Float64("lon", 0, "longitude of the search center")
	radiusFlag := cmd.Flags().Int("radius", 0, "search radius in meters (0–100000)")
	categoryFlag := cmd.Flags().StringSlice("category", nil, "Foursquare category ID(s) to filter by")
	limitFlag := cmd.Flags().Int("limit", 0, "max results (1–50, default 10)")
	_ = cmd.MarkFlagRequired("lat")
	_ = cmd.MarkFlagRequired("lon")

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := newClient(cfg, *keyFlag)
		if err != nil {
			return err
		}

		// Flatten category IDs: support both comma-separated and repeated flags.
		var categoryIDs []foursquare.CategoryID
		for _, v := range *categoryFlag {
			for _, id := range strings.Split(v, ",") {
				id = strings.TrimSpace(id)
				if id != "" {
					categoryIDs = append(categoryIDs, foursquare.CategoryID(id))
				}
			}
		}

		resp, err := client.Search(cmd.Context(), &foursquare.SearchRequest{
			Latitude:    *latFlag,
			Longitude:   *lonFlag,
			Radius:      *radiusFlag,
			CategoryIDs: categoryIDs,
			Limit:       *limitFlag,
		})
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(resp, "", "  ")
		cmd.Println(string(out))
		return nil
	}
	return cmd
}

// newGetPlaceCommand creates the get-place command.
func newGetPlaceCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get-place",
		Short:   "Get details for a place by fsq_place_id",
		GroupID: "api",
	}
	keyFlag := cmd.Flags().String("key", "", "Foursquare API key (overrides stored credentials)")
	idFlag := cmd.Flags().String("id", "", "Foursquare place ID (fsq_place_id)")
	_ = cmd.MarkFlagRequired("id")

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := newClient(cfg, *keyFlag)
		if err != nil {
			return err
		}
		resp, err := client.GetPlace(cmd.Context(), &foursquare.GetPlaceRequest{
			FSQPlaceID: *idFlag,
		})
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(resp, "", "  ")
		cmd.Println(string(out))
		return nil
	}
	return cmd
}

// newAuthCommand creates the auth command group with login/logout subcommands.
func newAuthCommand(cfg *config) *cobra.Command {
	auth := &cobra.Command{
		Use:     "auth",
		Short:   "Manage Foursquare API credentials",
		GroupID: "auth",
	}
	auth.AddCommand(newAuthLoginCommand(cfg), newAuthLogoutCommand(cfg))
	return auth
}

func newAuthLoginCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save a Foursquare API key to disk",
	}
	keyFlag := cmd.Flags().String("key", "", "Foursquare API key to store")
	_ = cmd.MarkFlagRequired("key")

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		if cfg.credentialStore == nil {
			return fmt.Errorf("no credential store configured")
		}
		creds := Credentials{APIKey: *keyFlag}
		if err := cfg.credentialStore.Write(creds); err != nil {
			return fmt.Errorf("save credentials: %w", err)
		}
		cmd.Println("Credentials saved.")
		return nil
	}
	return cmd
}

func newAuthLogoutCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored Foursquare credentials",
	}
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		if cfg.credentialStore == nil {
			return fmt.Errorf("no credential store configured")
		}
		if err := cfg.credentialStore.Clear(); err != nil {
			return fmt.Errorf("clear credentials: %w", err)
		}
		cmd.Println("Credentials removed.")
		return nil
	}
	return cmd
}

// newClient constructs a foursquare.Client from the key flag or the stored credentials.
func newClient(cfg *config, keyFlag string) (*foursquare.Client, error) {
	key := keyFlag
	if key == "" && cfg.credentialStore != nil {
		var creds Credentials
		if err := cfg.credentialStore.Read(&creds); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("no credentials found; run `foursquare auth login --key <KEY>` first")
			}
			return nil, fmt.Errorf("read credentials: %w", err)
		}
		key = creds.APIKey
	}
	if key == "" {
		return nil, fmt.Errorf("no API key: use --key or run `foursquare auth login`")
	}

	opts := []foursquare.Option{foursquare.WithAPIKey(key)}
	if cfg.httpClient != nil {
		opts = append(opts, foursquare.WithHTTPClient(cfg.httpClient))
	}
	return foursquare.NewClient(opts...), nil
}
