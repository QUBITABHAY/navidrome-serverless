package cmd

import (
	"context"

	"github.com/navidrome/navidrome/core"
	"github.com/navidrome/navidrome/core/agents"
	"github.com/navidrome/navidrome/core/artwork"
	"github.com/navidrome/navidrome/core/external"
	"github.com/navidrome/navidrome/core/ffmpeg"
	"github.com/navidrome/navidrome/core/lyrics"
	"github.com/navidrome/navidrome/core/matcher"
	"github.com/navidrome/navidrome/core/metrics"
	"github.com/navidrome/navidrome/core/playback"
	"github.com/navidrome/navidrome/core/playlists"
	"github.com/navidrome/navidrome/core/quickconnect"
	"github.com/navidrome/navidrome/core/scrobbler"
	"github.com/navidrome/navidrome/core/sonic"
	"github.com/navidrome/navidrome/core/stream"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/plugins"
	"github.com/navidrome/navidrome/scanner"
	"github.com/navidrome/navidrome/server/events"
	"github.com/navidrome/navidrome/server/nativeapi"
	"github.com/navidrome/navidrome/server/public"
	"github.com/navidrome/navidrome/server/subsonic"
)

// CreateNativeAPIRouterWithDS constructs the Native API router backed by the specified dataStore.
func CreateNativeAPIRouterWithDS(ctx context.Context, dataStore model.DataStore) *nativeapi.Router {
	share := core.NewShare(dataStore)
	uploader := artwork.NewUploader(dataStore)
	playlistsPlaylists := playlists.NewPlaylists(dataStore, uploader)
	insights := metrics.GetInstance(dataStore)
	broker := events.GetBroker()
	metricsMetrics := metrics.GetPrometheusInstance(dataStore)
	modelScanner := scanner.GetInstance(ctx, dataStore, broker, playlistsPlaylists, metricsMetrics)
	watcher := scanner.GetWatcher(dataStore, modelScanner)
	manager := plugins.GetManager(dataStore, broker, metricsMetrics)
	library := core.NewLibrary(dataStore, modelScanner, watcher, broker, manager)
	user := core.NewUser(dataStore, manager)
	maintenance := core.NewMaintenance(dataStore)
	agentsAgents := agents.GetAgents(dataStore, manager)
	matcherMatcher := matcher.New(dataStore)
	provider := external.NewProvider(dataStore, agentsAgents, matcherMatcher, broker)
	quickConnect := quickconnect.GetInstance()
	router := nativeapi.New(dataStore, share, playlistsPlaylists, insights, library, user, maintenance, manager, uploader, provider, quickConnect)
	return router
}

// CreateSubsonicAPIRouterWithDS constructs the Subsonic API router backed by the specified dataStore.
func CreateSubsonicAPIRouterWithDS(ctx context.Context, dataStore model.DataStore) *subsonic.Router {
	fileCache := artwork.GetImageCache()
	imageStore := artwork.GetImageStore()
	fFmpeg := ffmpeg.New()
	artworkArtwork := artwork.NewArtwork(dataStore, fileCache, imageStore, fFmpeg)
	transcodingCache := stream.GetTranscodingCache()
	mediaStreamer := stream.NewMediaStreamer(dataStore, fFmpeg, transcodingCache)
	share := core.NewShare(dataStore)
	archiver := core.NewArchiver(mediaStreamer, dataStore, share)
	players := core.NewPlayers(dataStore)
	broker := events.GetBroker()
	metricsMetrics := metrics.GetPrometheusInstance(dataStore)
	manager := plugins.GetManager(dataStore, broker, metricsMetrics)
	agentsAgents := agents.GetAgents(dataStore, manager)
	matcherMatcher := matcher.New(dataStore)
	provider := external.NewProvider(dataStore, agentsAgents, matcherMatcher, broker)
	uploader := artwork.NewUploader(dataStore)
	playlistsPlaylists := playlists.NewPlaylists(dataStore, uploader)
	modelScanner := scanner.GetInstance(ctx, dataStore, broker, playlistsPlaylists, metricsMetrics)
	playTracker := scrobbler.GetPlayTracker(dataStore, broker, manager)
	playbackServer := playback.GetInstance(dataStore)
	lyricsLyrics := lyrics.NewLyrics(dataStore, manager)
	transcodeDecider := stream.NewTranscodeDecider(dataStore, fFmpeg)
	sonicSonic := sonic.New(dataStore, manager, matcherMatcher)
	router := subsonic.New(dataStore, artworkArtwork, mediaStreamer, archiver, players, provider, modelScanner, broker, playlistsPlaylists, playTracker, share, playbackServer, metricsMetrics, lyricsLyrics, transcodeDecider, sonicSonic)
	return router
}

// CreatePublicRouterWithDS constructs the Public router backed by the specified dataStore.
func CreatePublicRouterWithDS(dataStore model.DataStore) *public.Router {
	fileCache := artwork.GetImageCache()
	imageStore := artwork.GetImageStore()
	fFmpeg := ffmpeg.New()
	artworkArtwork := artwork.NewArtwork(dataStore, fileCache, imageStore, fFmpeg)
	transcodingCache := stream.GetTranscodingCache()
	mediaStreamer := stream.NewMediaStreamer(dataStore, fFmpeg, transcodingCache)
	share := core.NewShare(dataStore)
	archiver := core.NewArchiver(mediaStreamer, dataStore, share)
	router := public.New(dataStore, artworkArtwork, mediaStreamer, share, archiver)
	return router
}
