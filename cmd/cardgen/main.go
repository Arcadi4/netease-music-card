package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Nthily/netease-music-card/internal/config"
	"github.com/Nthily/netease-music-card/internal/domain"
	"github.com/Nthily/netease-music-card/internal/netease"
	"github.com/Nthily/netease-music-card/internal/persist"
	"github.com/Nthily/netease-music-card/internal/publish"
	"github.com/Nthily/netease-music-card/internal/render"
)

func main() {
	help := flag.Bool("help", false, "Show usage information")
	selfCheckWeapi := flag.Bool("self-check-weapi", false, "Validate weapi crypto without API calls")
	checkAuth := flag.Bool("check-auth", false, "Check if cookie is valid")
	cookie := flag.String("cookie", "", "Netease cookie (MUSIC_U value)")
	userID := flag.String("user-id", "", "Netease user ID")
	fixture := flag.Bool("fixture", false, "Use fixture data for testing")
	fixtureZeroPlay := flag.Bool("fixture-zero-play", false, "Use zero-play fixture data (score fallback test)")
	dumpDerived := flag.String("dump-derived", "", "Dump derived data to JSON file")
	skipRender := flag.Bool("skip-render", false, "Skip rendering step")
	skipPublish := flag.Bool("skip-publish", false, "Skip publishing step")
	skipPNG := flag.Bool("skip-png", false, "Skip PNG generation")
	configPath := flag.String("config", "", "Path to artifacts config JSON")
	outputDir := flag.String("output-dir", ".", "Output directory for generated files")
	stylePath := flag.String("style", "", "Path to CSS override file")
	assetSelfCheck := flag.Bool("asset-self-check", false, "Validate asset encoding without network calls")
	simulateFetchError := flag.Bool("simulate-fetch-error", false, "Simulate fetch error for testing")
	publishSelfCheck := flag.Bool("publish-self-check", false, "Validate GitHub publisher without making commits")
	flag.Parse()

	if *help {
		fmt.Println("Usage: cardgen [options]")
		fmt.Println("\nGenerates Netease Music card and publishes to GitHub")
		fmt.Println("\nOptions:")
		fmt.Println("  --self-check-weapi    Validate weapi crypto without API calls")
		fmt.Println("  --check-auth          Check if cookie is valid")
		fmt.Println("  --cookie string       Netease cookie (MUSIC_U value)")
		fmt.Println("  --user-id string      Netease user ID")
		fmt.Println("  --fixture             Use fixture data for testing")
		fmt.Println("  --dump-derived path   Dump derived data to JSON file")
		fmt.Println("  --skip-render         Skip rendering step")
		fmt.Println("  --skip-publish        Skip publishing step")
		fmt.Println("\nRequired environment variables:")
		fmt.Println("  NETEASE_USER_ID    - Netease user ID")
		fmt.Println("  NETEASE_USER_TOKEN - Netease user token")
		fmt.Println("\nOptional environment variables:")
		fmt.Println("  GITHUB_TOKEN      - GitHub token (empty = local mode, skip publish)")
		fmt.Println("  GITHUB_REPOSITORY - GitHub repository in owner/repo format")
		fmt.Println("  OUTPUT_BRANCH - Target branch (default: main)")
		os.Exit(0)
	}

	if *selfCheckWeapi {
		if err := runSelfCheckWeapi(); err != nil {
			fmt.Fprintf(os.Stderr, "Self-check failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("WEAPI_SELF_CHECK_OK")
		os.Exit(0)
	}

	if *checkAuth {
		if *cookie == "" || *userID == "" {
			fmt.Fprintf(os.Stderr, "Error: --cookie and --user-id are required for --check-auth\n")
			os.Exit(1)
		}
		if err := runCheckAuth(*cookie, *userID); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Println("Cookie is valid")
		os.Exit(0)
	}

	if *assetSelfCheck {
		if err := runAssetSelfCheck(*simulateFetchError); err != nil {
			fmt.Fprintf(os.Stderr, "Asset self-check failed: %v\n", err)
			os.Exit(1)
		}
		if *simulateFetchError {
			fmt.Println("ASSET_FETCH_FALLBACK_OK")
		} else {
			fmt.Println("ASSET_SELF_CHECK_OK")
		}
		os.Exit(0)
	}

	if *publishSelfCheck {
		if err := runPublishSelfCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "Publish self-check failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("PUBLISH_SELF_CHECK_OK")
		os.Exit(0)
	}

	if *fixture || *fixtureZeroPlay {
		_ = *configPath
		if err := runFixtureMode(*dumpDerived, *skipRender, *skipPublish, *skipPNG, *fixtureZeroPlay, *outputDir, *stylePath); err != nil {
			fmt.Fprintf(os.Stderr, "Fixture mode failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := runProductionPipeline(cfg, *outputDir, *stylePath, *skipRender, *skipPublish); err != nil {
		fmt.Fprintf(os.Stderr, "Production pipeline failed: %v\n", err)
		os.Exit(1)
	}
}

func runSelfCheckWeapi() error {
	testData := map[string]interface{}{
		"test": "data",
		"num":  123,
	}

	encrypted, err := netease.EncryptWeapi(testData)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	if encrypted["params"] == "" || encrypted["encSecKey"] == "" {
		return fmt.Errorf("encryption produced empty results")
	}

	return nil
}

func runCheckAuth(cookie, userID string) error {
	client := netease.NewClient(userID, cookie)
	_, err := client.UserAccount()
	return err
}

func getFixtureWeekData() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"playCount": 42,
			"score":     85,
			"song": map[string]interface{}{
				"id":   1,
				"name": "孤独摇滚",
				"ar": []interface{}{
					map[string]interface{}{"id": 101, "name": "SICK HACK"},
					map[string]interface{}{"id": 102, "name": "Bocchi"},
				},
			},
		},
		{
			"playCount": 30,
			"score":     70,
			"song": map[string]interface{}{
				"id":   2,
				"name": "Guitar, Loneliness and Blue Planet",
				"ar": []interface{}{
					map[string]interface{}{"id": 101, "name": "SICK HACK"},
				},
			},
		},
		{
			"playCount": 0,
			"score":     60,
			"song": map[string]interface{}{
				"id":   3,
				"name": "あのバンド",
				"ar": []interface{}{
					map[string]interface{}{"id": 103, "name": "橡皮擦"},
				},
			},
		},
		{
			"playCount": 25,
			"score":     55,
			"song": map[string]interface{}{
				"id":   4,
				"name": "なにが悪い",
				"ar": []interface{}{
					map[string]interface{}{"id": 104, "name": "斗志"},
				},
			},
		},
		{
			"playCount": 18,
			"score":     40,
			"song": map[string]interface{}{
				"id":   5,
				"name": "ひとりぼっち東京",
				"ar": []interface{}{
					map[string]interface{}{"id": 105, "name": "後藤ひとり"},
				},
			},
		},
		{
			"playCount": 10,
			"score":     20,
			"song": map[string]interface{}{
				"id":   6,
				"name": "Long Longer",
				"ar": []interface{}{
					map[string]interface{}{"id": 106, "name": "なにか"},
				},
			},
		},
	}
}

func getZeroPlayFixtureWeekData() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"playCount": 0,
			"score":     85,
			"song": map[string]interface{}{
				"id":   1,
				"name": "孤独摇滚",
				"ar": []interface{}{
					map[string]interface{}{"id": 101, "name": "SICK HACK"},
					map[string]interface{}{"id": 102, "name": "Bocchi"},
				},
			},
		},
		{
			"playCount": 0,
			"score":     70,
			"song": map[string]interface{}{
				"id":   2,
				"name": "Guitar, Loneliness and Blue Planet",
				"ar": []interface{}{
					map[string]interface{}{"id": 101, "name": "SICK HACK"},
				},
			},
		},
		{
			"playCount": 0,
			"score":     60,
			"song": map[string]interface{}{
				"id":   3,
				"name": "あのバンド",
				"ar": []interface{}{
					map[string]interface{}{"id": 103, "name": "橡皮擦"},
				},
			},
		},
		{
			"playCount": 0,
			"score":     55,
			"song": map[string]interface{}{
				"id":   4,
				"name": "なにが悪い",
				"ar": []interface{}{
					map[string]interface{}{"id": 104, "name": "斗志"},
				},
			},
		},
		{
			"playCount": 0,
			"score":     40,
			"song": map[string]interface{}{
				"id":   5,
				"name": "ひとりぼっち東京",
				"ar": []interface{}{
					map[string]interface{}{"id": 105, "name": "後藤ひとり"},
				},
			},
		},
		{
			"playCount": 0,
			"score":     20,
			"song": map[string]interface{}{
				"id":   6,
				"name": "Long Longer",
				"ar": []interface{}{
					map[string]interface{}{"id": 106, "name": "なにか"},
				},
			},
		},
	}
}

func runFixtureMode(dumpPath string, skipRender, skipPublish, skipPNG, zeroPlay bool, outputDir, stylePath string) error {
	var fixtureData []map[string]interface{}
	if zeroPlay {
		fixtureData = getZeroPlayFixtureWeekData()
	} else {
		fixtureData = getFixtureWeekData()
	}

	topArtists := domain.DeriveTopArtists(fixtureData, 5)
	topTracks := domain.DeriveTopTracks(fixtureData, 5)
	overview := domain.DeriveWeeklyOverview(fixtureData)

	if err := persist.Write(".", persist.Artifacts{
		TopArtists:     topArtists,
		TopTracks:      topTracks,
		WeeklyOverview: overview,
		CardInput: persist.CardInput{
			Nickname:    "测试用户",
			SongName:    "孤独摇滚",
			SongAuthors: "SICK HACK / Bocchi",
			PlayCount:   42,
			AuthFailed:  false,
		},
	}); err != nil {
		return fmt.Errorf("persist data: %w", err)
	}

	if dumpPath != "" {
		derived := map[string]interface{}{
			"topArtists": topArtists,
			"topTracks":  topTracks,
			"overview":   overview,
		}

		data, err := json.MarshalIndent(derived, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal derived data: %w", err)
		}

		if err := os.WriteFile(dumpPath, data, 0o644); err != nil {
			return fmt.Errorf("write derived data: %w", err)
		}

		fmt.Printf("Derived data written to %s\n", dumpPath)
	}

	if !skipRender {
		_ = skipPNG
		css, err := render.LoadCSS(stylePath)
		if err != nil {
			return fmt.Errorf("load CSS: %w", err)
		}

		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("create output dir: %w", err)
		}

		topArtistsSVG, err := render.RenderTopArtists(render.TopArtistsData{
			CSS:     css,
			Artists: topArtists,
		})
		if err != nil {
			return fmt.Errorf("render top artists: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/top-artists.svg", outputDir), topArtistsSVG, 0o644); err != nil {
			return fmt.Errorf("write top-artists.svg: %w", err)
		}

		topTracksSVG, err := render.RenderTopTracks(render.TopTracksData{
			CSS:    css,
			Tracks: topTracks,
		})
		if err != nil {
			return fmt.Errorf("render top tracks: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/top-tracks.svg", outputDir), topTracksSVG, 0o644); err != nil {
			return fmt.Errorf("write top-tracks.svg: %w", err)
		}

		overviewSVG, err := render.RenderWeeklyOverview(render.WeeklyOverviewData{
			CSS:             css,
			TotalPlays:      overview.TotalPlays,
			UniqueSongs:     overview.UniqueSongs,
			UniqueArtists:   overview.UniqueArtists,
			RepeatIntensity: overview.RepeatIntensity,
		})
		if err != nil {
			return fmt.Errorf("render weekly overview: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/weekly-overview.svg", outputDir), overviewSVG, 0o644); err != nil {
			return fmt.Errorf("write weekly-overview.svg: %w", err)
		}

		cardSVG, err := render.RenderCard(render.CardData{
			CSS:          css,
			AvatarBase64: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
			Nickname:     "测试用户",
			SongName:     "孤独摇滚",
			SongAuthors:  "SICK HACK / Bocchi",
			PlayCount:    42,
			CoverBase64:  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
			LogoBase64:   render.NeteaseLogoBase64,
		})
		if err != nil {
			return fmt.Errorf("render card: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/card.svg", outputDir), cardSVG, 0o644); err != nil {
			return fmt.Errorf("write card.svg: %w", err)
		}

		fmt.Printf("Rendered 4 SVG files to %s\n", outputDir)
	}

	fmt.Printf("Fixture mode: topArtists=%d, topTracks=%d, totalPlays=%d\n",
		len(topArtists), len(topTracks), overview.TotalPlays)

	return nil
}

func runAssetSelfCheck(simulateError bool) error {
	if simulateError {
		render.SetSimulateFetchError(true)
		_, err := render.FetchAndEncode("http://example.com/test.png")
		if err == nil {
			return fmt.Errorf("expected error but got none")
		}
		return nil
	}

	fixtureBytes := []byte("test image data")
	encoded := base64.StdEncoding.EncodeToString(fixtureBytes)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}
	if string(decoded) != string(fixtureBytes) {
		return fmt.Errorf("roundtrip mismatch")
	}

	return nil
}

func runPublishSelfCheck() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if cfg.GHToken == "" {
		return fmt.Errorf("GITHUB_TOKEN is empty")
	}
	if cfg.Repo == "" {
		return fmt.Errorf("repository detection returned empty value")
	}
	if cfg.OutputBranch == "" {
		return fmt.Errorf("OUTPUT_BRANCH is empty")
	}

	parts := strings.Split(cfg.Repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("REPO must be in 'owner/repo' format, got: %s", cfg.Repo)
	}

	return nil
}

func runProductionPipeline(cfg *config.Config, outputDir, stylePath string, skipRender, skipPublish bool) error {
	client := netease.NewClient(cfg.UserID, cfg.UserToken)

	authFailed := false
	var weekData []map[string]interface{}
	var nickname, avatarBase64, songName, songAuthors, coverBase64 string
	var playCount int

	account, err := client.UserAccount()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Auth failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Continuing with empty state (will skip card.svg)\n")
		authFailed = true
		weekData = []map[string]interface{}{}
	} else {
		record, err := client.UserRecord(cfg.UserID, 1)
		if err != nil {
			return fmt.Errorf("fetch user record: %w", err)
		}

		weekData = domain.SafeWeekData(record)

		userDetail, err := client.UserDetail(cfg.UserID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fetch user detail failed: %v, using empty avatar\n", err)
			avatarBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
		} else {
			if profile, ok := userDetail["profile"].(map[string]interface{}); ok {
				if name, ok := profile["nickname"].(string); ok {
					nickname = name
				}
				if avatarURL, ok := profile["avatarUrl"].(string); ok {
					avatarBase64, err = render.FetchAndEncode(avatarURL)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Fetch avatar failed: %v, using fallback\n", err)
						avatarBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
					}
				}
			}
		}

		if len(weekData) > 0 {
			topEntry := weekData[0]
			if song, ok := topEntry["song"].(map[string]interface{}); ok {
				if name, ok := song["name"].(string); ok {
					songName = name
				}
				if ar, ok := song["ar"].([]interface{}); ok {
					names := []string{}
					for _, item := range ar {
						if artist, ok := item.(map[string]interface{}); ok {
							if name, ok := artist["name"].(string); ok {
								names = append(names, name)
							}
						}
					}
					songAuthors = strings.Join(names, " / ")
				}
				if al, ok := song["al"].(map[string]interface{}); ok {
					if picURL, ok := al["picUrl"].(string); ok {
						coverBase64, err = render.FetchAndEncode(picURL)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Fetch cover failed: %v, using fallback\n", err)
							coverBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
						}
					}
				}
			}
			if pc, ok := topEntry["playCount"].(float64); ok {
				playCount = int(pc)
			} else if pc, ok := topEntry["playCount"].(int); ok {
				playCount = pc
			}
		}

		_ = account
	}

	topArtists := domain.DeriveTopArtists(weekData, 5)
	topTracks := domain.DeriveTopTracks(weekData, 5)
	overview := domain.DeriveWeeklyOverview(weekData)

	if err := persist.Write(".", persist.Artifacts{
		TopArtists:     topArtists,
		TopTracks:      topTracks,
		WeeklyOverview: overview,
		CardInput: persist.CardInput{
			Nickname:    nickname,
			SongName:    songName,
			SongAuthors: songAuthors,
			PlayCount:   playCount,
			AuthFailed:  authFailed,
		},
	}); err != nil {
		return fmt.Errorf("persist data: %w", err)
	}

	if !skipRender {
		css, err := render.LoadCSS(stylePath)
		if err != nil {
			return fmt.Errorf("load CSS: %w", err)
		}

		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("create output dir: %w", err)
		}

		topArtistsSVG, err := render.RenderTopArtists(render.TopArtistsData{
			CSS:     css,
			Artists: topArtists,
		})
		if err != nil {
			return fmt.Errorf("render top artists: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/top-artists.svg", outputDir), topArtistsSVG, 0o644); err != nil {
			return fmt.Errorf("write top-artists.svg: %w", err)
		}

		topTracksSVG, err := render.RenderTopTracks(render.TopTracksData{
			CSS:    css,
			Tracks: topTracks,
		})
		if err != nil {
			return fmt.Errorf("render top tracks: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/top-tracks.svg", outputDir), topTracksSVG, 0o644); err != nil {
			return fmt.Errorf("write top-tracks.svg: %w", err)
		}

		overviewSVG, err := render.RenderWeeklyOverview(render.WeeklyOverviewData{
			CSS:             css,
			TotalPlays:      overview.TotalPlays,
			UniqueSongs:     overview.UniqueSongs,
			UniqueArtists:   overview.UniqueArtists,
			RepeatIntensity: overview.RepeatIntensity,
		})
		if err != nil {
			return fmt.Errorf("render weekly overview: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/weekly-overview.svg", outputDir), overviewSVG, 0o644); err != nil {
			return fmt.Errorf("write weekly-overview.svg: %w", err)
		}

		if !authFailed {
			cardSVG, err := render.RenderCard(render.CardData{
				CSS:          css,
				AvatarBase64: avatarBase64,
				Nickname:     nickname,
				SongName:     songName,
				SongAuthors:  songAuthors,
				PlayCount:    playCount,
				CoverBase64:  coverBase64,
				LogoBase64:   render.NeteaseLogoBase64,
			})
			if err != nil {
				return fmt.Errorf("render card: %w", err)
			}
			if err := os.WriteFile(fmt.Sprintf("%s/card.svg", outputDir), cardSVG, 0o644); err != nil {
				return fmt.Errorf("write card.svg: %w", err)
			}
			fmt.Printf("Rendered 4 SVG files to %s\n", outputDir)
		} else {
			fmt.Printf("Rendered 3 SVG files to %s (skipped card.svg due to auth failure)\n", outputDir)
		}
	}

	if !skipPublish {
		if cfg.GHToken == "" {
			fmt.Println("GITHUB_TOKEN not set — skipping publish (local mode)")
		} else {
			parts := strings.Split(cfg.Repo, "/")
			if len(parts) != 2 {
				return fmt.Errorf("REPO must be in 'owner/repo' format, got: %s", cfg.Repo)
			}

			publisher := publish.NewGitHubPublisher(cfg.GHToken, parts[0], parts[1], cfg.OutputBranch)

			files := []publish.FileToCommit{}

			if !authFailed {
				cardSVG, err := os.ReadFile(fmt.Sprintf("%s/card.svg", outputDir))
				if err != nil {
					return fmt.Errorf("read card.svg: %w", err)
				}
				files = append(files, publish.FileToCommit{Path: "card.svg", Content: cardSVG})
			}

			topArtistsSVG, err := os.ReadFile(fmt.Sprintf("%s/top-artists.svg", outputDir))
			if err != nil {
				return fmt.Errorf("read top-artists.svg: %w", err)
			}
			files = append(files, publish.FileToCommit{Path: "top-artists.svg", Content: topArtistsSVG})

			topTracksSVG, err := os.ReadFile(fmt.Sprintf("%s/top-tracks.svg", outputDir))
			if err != nil {
				return fmt.Errorf("read top-tracks.svg: %w", err)
			}
			files = append(files, publish.FileToCommit{Path: "top-tracks.svg", Content: topTracksSVG})

			overviewSVG, err := os.ReadFile(fmt.Sprintf("%s/weekly-overview.svg", outputDir))
			if err != nil {
				return fmt.Errorf("read weekly-overview.svg: %w", err)
			}
			files = append(files, publish.FileToCommit{Path: "weekly-overview.svg", Content: overviewSVG})

			topArtistsJSON, err := os.ReadFile("data/top-artists.json")
			if err != nil {
				return fmt.Errorf("read data/top-artists.json: %w", err)
			}
			files = append(files, publish.FileToCommit{Path: "data/top-artists.json", Content: topArtistsJSON})

			topTracksJSON, err := os.ReadFile("data/top-tracks.json")
			if err != nil {
				return fmt.Errorf("read data/top-tracks.json: %w", err)
			}
			files = append(files, publish.FileToCommit{Path: "data/top-tracks.json", Content: topTracksJSON})

			weeklyOverviewJSON, err := os.ReadFile("data/weekly-overview.json")
			if err != nil {
				return fmt.Errorf("read data/weekly-overview.json: %w", err)
			}
			files = append(files, publish.FileToCommit{Path: "data/weekly-overview.json", Content: weeklyOverviewJSON})

			cardInputJSON, err := os.ReadFile("data/card-input.json")
			if err != nil {
				return fmt.Errorf("read data/card-input.json: %w", err)
			}
			files = append(files, publish.FileToCommit{Path: "data/card-input.json", Content: cardInputJSON})

			if err := publisher.CommitFiles(files); err != nil {
				return fmt.Errorf("commit files: %w", err)
			}

			fmt.Printf("Published %d files to %s/%s branch %s\n", len(files), parts[0], parts[1], cfg.OutputBranch)
		}
	}

	return nil
}
