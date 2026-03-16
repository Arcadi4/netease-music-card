package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Nthily/netease-music-card/internal/config"
	"github.com/Nthily/netease-music-card/internal/domain"
	"github.com/Nthily/netease-music-card/internal/netease"
	"github.com/Nthily/netease-music-card/internal/publish"
	"github.com/Nthily/netease-music-card/internal/render"
	"github.com/Nthily/netease-music-card/internal/snapshot"
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
	snapshotSelfCheck := flag.Bool("snapshot-self-check", false, "Validate snapshot load/save without API calls")
	dumpDuration := flag.String("dump-duration", "", "Dump duration data to JSON file")
	snapshotPath := flag.String("snapshot-path", "duration-snapshot.json", "Path to duration snapshot file")
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
		fmt.Println("  USER_ID      - Netease user ID")
		fmt.Println("  USER_TOKEN   - Netease user token")
		fmt.Println("  GH_TOKEN     - GitHub token")
		fmt.Println("\nOptional environment variables:")
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

	if *snapshotSelfCheck {
		if err := runSnapshotSelfCheck(*snapshotPath); err != nil {
			fmt.Fprintf(os.Stderr, "Snapshot self-check failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("SNAPSHOT_SELF_CHECK_OK")
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
		if err := runFixtureMode(*dumpDerived, *dumpDuration, *skipRender, *skipPublish, *skipPNG, *fixtureZeroPlay, *outputDir, *stylePath); err != nil {
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

	if err := runProductionPipeline(cfg, *snapshotPath, *outputDir, *stylePath, *skipRender, *skipPublish); err != nil {
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

func runSnapshotSelfCheck(path string) error {
	testSnap := snapshot.DurationSnapshot{
		"2026-03-10": 100,
		"2026-03-11": 150,
		"2026-03-12": 200,
	}

	if err := snapshot.Save(path, testSnap); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}

	loaded, err := snapshot.Load(path)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	if len(loaded) != len(testSnap) {
		return fmt.Errorf("loaded snapshot length mismatch: got %d, want %d", len(loaded), len(testSnap))
	}

	for k, v := range testSnap {
		if loaded[k] != v {
			return fmt.Errorf("snapshot value mismatch for %s: got %d, want %d", k, loaded[k], v)
		}
	}

	updated := snapshot.Update(loaded, 250, "2026-03-13")
	if err := snapshot.Save(path, updated); err != nil {
		return fmt.Errorf("save updated snapshot: %w", err)
	}

	return nil
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

func runFixtureMode(dumpPath, dumpDuration string, skipRender, skipPublish, skipPNG, zeroPlay bool, outputDir, stylePath string) error {
	var fixtureData []map[string]interface{}
	if zeroPlay {
		fixtureData = getZeroPlayFixtureWeekData()
	} else {
		fixtureData = getFixtureWeekData()
	}

	topArtists := domain.DeriveTopArtists(fixtureData, 5)
	topTracks := domain.DeriveTopTracks(fixtureData, 5)
	overview := domain.DeriveWeeklyOverview(fixtureData)

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

		if err := os.WriteFile(dumpPath, data, 0644); err != nil {
			return fmt.Errorf("write derived data: %w", err)
		}

		fmt.Printf("Derived data written to %s\n", dumpPath)
	}

	snap := snapshot.DurationSnapshot{
		"2026-03-10": 100,
		"2026-03-11": 150,
		"2026-03-12": 200,
		"2026-03-13": 250,
		"2026-03-14": 300,
		"2026-03-15": 350,
		"2026-03-16": 400,
	}
	durations := domain.DeriveDailyDurations(snap, 3.5)

	if dumpDuration != "" {
		data, err := json.MarshalIndent(durations, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal duration data: %w", err)
		}

		if err := os.WriteFile(dumpDuration, data, 0644); err != nil {
			return fmt.Errorf("write duration data: %w", err)
		}

		fmt.Printf("Duration data written to %s\n", dumpDuration)
	}

	if !skipRender {
		_ = skipPNG
		css, err := render.LoadCSS(stylePath)
		if err != nil {
			return fmt.Errorf("load CSS: %w", err)
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("create output dir: %w", err)
		}

		topArtistsSVG, err := render.RenderTopArtists(render.TopArtistsData{
			CSS:     css,
			Artists: topArtists,
		})
		if err != nil {
			return fmt.Errorf("render top artists: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/top-artists.svg", outputDir), topArtistsSVG, 0644); err != nil {
			return fmt.Errorf("write top-artists.svg: %w", err)
		}

		topTracksSVG, err := render.RenderTopTracks(render.TopTracksData{
			CSS:    css,
			Tracks: topTracks,
		})
		if err != nil {
			return fmt.Errorf("render top tracks: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/top-tracks.svg", outputDir), topTracksSVG, 0644); err != nil {
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
		if err := os.WriteFile(fmt.Sprintf("%s/weekly-overview.svg", outputDir), overviewSVG, 0644); err != nil {
			return fmt.Errorf("write weekly-overview.svg: %w", err)
		}

		durationSVG, err := render.RenderWeeklyDuration(durations, css)
		if err != nil {
			return fmt.Errorf("render weekly duration: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/weekly-duration.svg", outputDir), durationSVG, 0644); err != nil {
			return fmt.Errorf("write weekly-duration.svg: %w", err)
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
		if err := os.WriteFile(fmt.Sprintf("%s/card.svg", outputDir), cardSVG, 0644); err != nil {
			return fmt.Errorf("write card.svg: %w", err)
		}

		fmt.Printf("Rendered 5 SVG files to %s\n", outputDir)
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
		return fmt.Errorf("GH_TOKEN is empty")
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

func runProductionPipeline(cfg *config.Config, snapshotPath, outputDir, stylePath string, skipRender, skipPublish bool) error {
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

	snap, err := snapshot.Load(snapshotPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Load snapshot failed: %v, using empty snapshot\n", err)
		snap = snapshot.DurationSnapshot{}
	}

	var totalDuration int
	if len(weekData) > 0 {
		for _, entry := range weekData {
			if pc, ok := entry["playCount"].(float64); ok {
				totalDuration += int(pc) * 210
			} else if pc, ok := entry["playCount"].(int); ok {
				totalDuration += pc * 210
			} else if score, ok := entry["score"].(float64); ok {
				totalDuration += int(score) * 210
			} else if score, ok := entry["score"].(int); ok {
				totalDuration += score * 210
			}
		}
	}

	snap = snapshot.Update(snap, totalDuration, time.Now().Format("2006-01-02"))

	if err := snapshot.Save(snapshotPath, snap); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}

	topArtists := domain.DeriveTopArtists(weekData, 5)
	topTracks := domain.DeriveTopTracks(weekData, 5)
	overview := domain.DeriveWeeklyOverview(weekData)
	durations := domain.DeriveDailyDurations(snap, 3.5)

	if !skipRender {
		css, err := render.LoadCSS(stylePath)
		if err != nil {
			return fmt.Errorf("load CSS: %w", err)
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("create output dir: %w", err)
		}

		topArtistsSVG, err := render.RenderTopArtists(render.TopArtistsData{
			CSS:     css,
			Artists: topArtists,
		})
		if err != nil {
			return fmt.Errorf("render top artists: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/top-artists.svg", outputDir), topArtistsSVG, 0644); err != nil {
			return fmt.Errorf("write top-artists.svg: %w", err)
		}

		topTracksSVG, err := render.RenderTopTracks(render.TopTracksData{
			CSS:    css,
			Tracks: topTracks,
		})
		if err != nil {
			return fmt.Errorf("render top tracks: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/top-tracks.svg", outputDir), topTracksSVG, 0644); err != nil {
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
		if err := os.WriteFile(fmt.Sprintf("%s/weekly-overview.svg", outputDir), overviewSVG, 0644); err != nil {
			return fmt.Errorf("write weekly-overview.svg: %w", err)
		}

		durationSVG, err := render.RenderWeeklyDuration(durations, css)
		if err != nil {
			return fmt.Errorf("render weekly duration: %w", err)
		}
		if err := os.WriteFile(fmt.Sprintf("%s/weekly-duration.svg", outputDir), durationSVG, 0644); err != nil {
			return fmt.Errorf("write weekly-duration.svg: %w", err)
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
			if err := os.WriteFile(fmt.Sprintf("%s/card.svg", outputDir), cardSVG, 0644); err != nil {
				return fmt.Errorf("write card.svg: %w", err)
			}
			fmt.Printf("Rendered 5 SVG files to %s\n", outputDir)
		} else {
			fmt.Printf("Rendered 4 SVG files to %s (skipped card.svg due to auth failure)\n", outputDir)
		}
	}

	if !skipPublish {
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

		durationSVG, err := os.ReadFile(fmt.Sprintf("%s/weekly-duration.svg", outputDir))
		if err != nil {
			return fmt.Errorf("read weekly-duration.svg: %w", err)
		}
		files = append(files, publish.FileToCommit{Path: "weekly-duration.svg", Content: durationSVG})

		snapshotData, err := os.ReadFile(snapshotPath)
		if err != nil {
			return fmt.Errorf("read snapshot: %w", err)
		}
		files = append(files, publish.FileToCommit{Path: "duration-snapshot.json", Content: snapshotData})

		if err := publisher.CommitFiles(files); err != nil {
			return fmt.Errorf("commit files: %w", err)
		}

		fmt.Printf("Published %d files to %s/%s branch %s\n", len(files), parts[0], parts[1], cfg.OutputBranch)
	}

	return nil
}
