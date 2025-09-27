package qbtapi

import (
	"context"
	"fmt"
)

/*
	Application
	https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#application
*/

const (
	applicationAPIName = "app"
)

// GetApplicationVersion returns the application version.
// https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#get-application-version
func (c *Client) GetApplicationVersion(ctx context.Context) (version string, err error) {
	req, err := c.requestBuild(ctx, "GET", applicationAPIName, "version", nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(ctx, req, &version, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// GetAPIVersion returns the WebAPI version.
// https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#get-api-version
func (c *Client) GetAPIVersion(ctx context.Context) (version string, err error) {
	req, err := c.requestBuild(ctx, "GET", applicationAPIName, "webapiVersion", nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(ctx, req, &version, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// BuildInfo contains all the compilation informations of a given remote server
type BuildInfo struct {
	QT         string `json:"qt"`         // QT version
	LibTorrent string `json:"libtorrent"` // libtorrent version
	Boost      string `json:"boost"`      // Boost version
	OpenSSL    string `json:"openssl"`    // OpenSSL version
	Bitness    int    `json:"bitness"`    // Application bitness (e.g. 64-bit)
}

// GetBuildInfo returns compilation informations of target server.
// https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#get-build-info
func (c *Client) GetBuildInfo(ctx context.Context) (infos BuildInfo, err error) {
	req, err := c.requestBuild(ctx, "GET", applicationAPIName, "buildInfo", nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(ctx, req, &infos, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// Shutdown stops the remote server.
// https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#shutdown-application
func (c *Client) Shutdown(ctx context.Context) (err error) {
	req, err := c.requestBuild(ctx, "POST", applicationAPIName, "shutdown", nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(ctx, req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// ApplicationPreferences references all the user preferences within the remote server.
// When getting preferences, all fields should be instanciated. When setting preferences only
// non nil pointers will be pass to the final payload (targeted update).
// See https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#get-application-preferences
// and https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#set-application-preferences
type ApplicationPreferences struct {
	AddTrackers                        *string      `json:"add_trackers"`              // List of trackers to add to new torrent
	AddTrackersEnabled                 *bool        `json:"add_trackers_enabled"`      // Enable automatic adding of trackers to new torrents
	AltDlLimit                         *int         `json:"alt_dl_limit"`              // Alternative global download speed limit in KiB/s [TODO]
	AltUpLimit                         *int         `json:"alt_up_limit"`              // Alternative global upload speed limit in KiB/s [TODO]
	AlternativeWebuiEnabled            *bool        `json:"alternative_webui_enabled"` // True if an alternative WebUI should be used
	AlternativeWebuiPath               *string      `json:"alternative_webui_path"`    // File path to the alternative WebUI
	AnnounceIP                         *string      `json:"announce_ip"`               // TODO
	AnnounceToAllTiers                 *bool        `json:"announce_to_all_tiers"`     // True always announce to all tiers
	AnnounceToAllTrackers              *bool        `json:"announce_to_all_trackers"`  // True always announce to all trackers in a tier
	AnonymousMode                      *bool        `json:"anonymous_mode"`            // If true anonymous mode will be enabled; read more here; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	AsyncIOThreads                     *int         `json:"async_io_threads"`          // Number of asynchronous I/O threads
	AutoDeleteMode                     *int         `json:"auto_delete_mode"`
	AutoTMMEnabled                     *bool        `json:"auto_tmm_enabled"`
	AutorunEnabled                     *bool        `json:"autorun_enabled"`                      // True if external program should be run after torrent has finished downloading
	AutorunProgram                     *string      `json:"autorun_program"`                      // Program path/name/arguments to run if autorun_enabled is enabled; path is separated by slashes; you can use %f and %n arguments, which will be expanded by qBittorent as path_to_torrent_file and torrent_name (from the GUI; not the .torrent file name) respectively
	BannedIPs                          *string      `json:"banned_IPs"`                           // List of banned IPs
	BittorrentProtocol                 *int         `json:"bittorrent_protocol"`                  // Bittorrent Protocol to use
	BypassAuthSubnetWhitelist          *string      `json:"bypass_auth_subnet_whitelist"`         // (White)list of ipv4/ipv6 subnets for which webui authentication should be bypassed; list entries are separated by commas
	BypassAuthSubnetWhitelistEnabled   *bool        `json:"bypass_auth_subnet_whitelist_enabled"` // True if webui authentication should be bypassed for clients whose ip resides within (at least) one of the subnets on the whitelist
	BypassLocalAuth                    *bool        `json:"bypass_local_auth"`                    // True if authentication challenge for loopback address (127.0.0.1) should be disabled
	CategoryChangedTmmEnabled          *bool        `json:"category_changed_tmm_enabled"`         // True if torrent should be relocated when its Category's save path changes
	CheckingMemoryUse                  *int         `json:"checking_memory_use"`                  // Outstanding memory when checking torrents in MiB
	CreateSubfolderEnabled             *bool        `json:"create_subfolder_enabled"`
	CurrentInterfaceAddress            *string      `json:"current_interface_address"`             // IP Address to bind to. Empty String means All addresses.
	CurrentNetworkInterface            *string      `json:"current_network_interface"`             // Network Interface used
	DHT                                *bool        `json:"dht"`                                   // True if DHT is enabled
	DiskCache                          *int         `json:"disk_cache"`                            // Disk cache used in MiB
	DiskCacheTTL                       *int         `json:"disk_cache_ttl"`                        // Disk cache expiry interval in seconds
	DlLimit                            *int         `json:"dl_limit"`                              // Global download speed limit in KiB/s; -1 means no limit is applied
	DontCountSlowTorrents              *bool        `json:"dont_count_slow_torrents"`              // If true torrents w/o any activity (stalled ones) will not be counted towards max_active_* limits; see dont_count_slow_torrents for more information
	DyndnsDomain                       *string      `json:"dyndns_domain"`                         // Your DDNS domain name
	DyndnsEnabled                      *bool        `json:"dyndns_enabled"`                        // True if server DNS should be updated dynamically
	DyndnsPassword                     *string      `json:"dyndns_password"`                       // Password for DDNS service
	DyndnsService                      *int         `json:"dyndns_service"`                        // See list of possible values here below
	DyndnsUsername                     *string      `json:"dyndns_username"`                       // Username for DDNS service
	EmbeddedTrackerPort                *int         `json:"embedded_tracker_port"`                 // Port used for embedded tracker
	EnableCoalesceReadWrite            *bool        `json:"enable_coalesce_read_write"`            // True enables coalesce reads & writes
	EnableEmbeddedTracker              *bool        `json:"enable_embedded_tracker"`               // True enables embedded tracker
	EnableMultiConnectionsFromSameIP   *bool        `json:"enable_multi_connections_from_same_ip"` // True allows multiple connections from the same IP address
	EnableOsCache                      *bool        `json:"enable_os_cache"`                       // True enables os cache
	EnablePieceExtentAffinity          *bool        `json:"enable_piece_extent_affinity"`          // True if the advanced libtorrent option piece_extent_affinity is enabled
	EnableUploadSuggestions            *bool        `json:"enable_upload_suggestions"`             // True enables sending of upload piece suggestions
	Encryption                         *int         `json:"encryption"`                            // See list of possible values here below [TODO]
	ExportDir                          *string      `json:"export_dir"`                            // Path to directory to copy .torrent files to. Slashes are used as path separators
	ExportDirFin                       *string      `json:"export_dir_fin"`                        // Path to directory to copy .torrent files of completed downloads to. Slashes are used as path separators
	FilePoolSize                       *int         `json:"file_pool_size"`                        // File pool size
	IncompleteFilesExt                 *bool        `json:"incomplete_files_ext"`
	IPFilterEnabled                    *bool        `json:"ip_filter_enabled"`  // True if external IP filter should be enabled
	IPFilterPath                       *string      `json:"ip_filter_path"`     // Path to IP filter file (.dat, .p2p, .p2b files are supported); path is separated by slashes
	IPFilterTrackers                   *bool        `json:"ip_filter_trackers"` // True if IP filters are applied to trackers
	LimitLanPeers                      *bool        `json:"limit_lan_peers"`    // True if [du]l_limit should be applied to peers on the LAN
	LimitTCPOverhead                   *bool        `json:"limit_tcp_overhead"` // True if [du]l_limit should be applied to estimated TCP overhead (service data: e.g. packet headers)
	LimitUtpRate                       *bool        `json:"limit_utp_rate"`     // True if [du]l_limit should be applied to uTP connections; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	ListenPort                         *int         `json:"listen_port"`        // Port for incoming connections
	Locale                             *string      `json:"locale"`
	LSD                                *bool        `json:"lsd"`                            // True if Local Service Discovery is enabled
	MailNotificationAuthEnabled        *bool        `json:"mail_notification_auth_enabled"` // True if smtp server requires authentication
	MailNotificationEmail              *string      `json:"mail_notification_email"`        // e-mail to send notifications to
	MailNotificationEnabled            *bool        `json:"mail_notification_enabled"`      // True if e-mail notification should be enabled
	MailNotificationPassword           *string      `json:"mail_notification_password"`     // Password for smtp authentication
	MailNotificationSender             *string      `json:"mail_notification_sender"`       // e-mail where notifications should originate from
	MailNotificationSMTP               *string      `json:"mail_notification_smtp"`         // smtp server for e-mail notifications
	MailNotificationSslEnabled         *bool        `json:"mail_notification_ssl_enabled"`  // True if smtp server requires SSL connection
	MailNotificationUsername           *string      `json:"mail_notification_username"`     // Username for smtp authentication
	MaxActiveDownloads                 *int         `json:"max_active_downloads"`           // Maximum number of active simultaneous downloads
	MaxActiveTorrents                  *int         `json:"max_active_torrents"`            // Maximum number of active simultaneous downloads and uploads
	MaxActiveUploads                   *int         `json:"max_active_uploads"`             // Maximum number of active simultaneous uploads
	MaxConnec                          *int         `json:"max_connec"`                     // Maximum global number of simultaneous connections
	MaxConnecPerTorrent                *int         `json:"max_connec_per_torrent"`         // Maximum number of simultaneous connections per torrent
	MaxRatio                           *int         `json:"max_ratio"`                      // Get the global share ratio limit
	MaxRatioAct                        *int         `json:"max_ratio_act"`                  // Action performed when a torrent reaches the maximum share ratio.
	MaxRatioEnabled                    *bool        `json:"max_ratio_enabled"`              // True if share ratio limit is enabled
	MaxSeedingTime                     *int         `json:"max_seeding_time"`               // Number of minutes to seed a torrent
	MaxSeedingTimeEnabled              *bool        `json:"max_seeding_time_enabled"`       // True enables max seeding time
	MaxUploads                         *int         `json:"max_uploads"`                    // Maximum number of upload slots
	MaxUploadsPerTorrent               *int         `json:"max_uploads_per_torrent"`        // Maximum number of upload slots per torrent
	OutgoingPortsMax                   *int         `json:"outgoing_ports_max"`             // Maximal outgoing port (0: Disabled)
	OutgoingPortsMin                   *int         `json:"outgoing_ports_min"`             // Minimal outgoing port (0: Disabled)
	PeX                                *bool        `json:"pex"`                            // True if PeX is enabled
	PreallocateAll                     *bool        `json:"preallocate_all"`
	ProxyAuthEnabled                   *bool        `json:"proxy_auth_enabled"`                  // True proxy requires authentication; doesn't apply to SOCKS4 proxies
	ProxyIP                            *string      `json:"proxy_ip"`                            // Proxy IP address or domain name
	ProxyPassword                      *string      `json:"proxy_password"`                      // Password for proxy authentication
	ProxyPeerConnections               *bool        `json:"proxy_peer_connections"`              // True if peer and web seed connections should be proxified; this option will have any effect only in qBittorent built against libtorrent version 0.16.X and higher
	ProxyPort                          *int         `json:"proxy_port"`                          // Proxy port
	ProxyTorrentsOnly                  *bool        `json:"proxy_torrents_only"`                 // True if proxy is only used for torrents
	ProxyType                          *int         `json:"proxy_type"`                          // See list of possible values here below
	ProxyUsername                      *string      `json:"proxy_username"`                      // Username for proxy authentication
	QueueingEnabled                    *bool        `json:"queueing_enabled"`                    // True if torrent queuing is enabled
	RandomPort                         *bool        `json:"random_port"`                         // True if the port is randomly selected
	RecheckCompletedTorrents           *bool        `json:"recheck_completed_torrents"`          // True rechecks torrents on completion
	ResolvePeerCountries               *bool        `json:"resolve_peer_countries"`              // True resolves peer countries
	RSSAutoDownloadingEnabled          *bool        `json:"rss_auto_downloading_enabled"`        // Enable auto-downloading of torrents from the RSS feeds
	RSSDownloadRepackProperEpisodes    *bool        `json:"rss_download_repack_proper_episodes"` // Enable downloading of repack/proper Episodes
	RSSMaxArticlesPerFeed              *int         `json:"rss_max_articles_per_feed"`           // Max stored articles per RSS feed
	RSSProcessingEnabled               *bool        `json:"rss_processing_enabled"`              // Enable processing of RSS feeds
	RSSRefreshInterval                 *int         `json:"rss_refresh_interval"`                // RSS refresh interval
	RSSSmartEpisodeFilters             *string      `json:"rss_smart_episode_filters"`           // List of RSS Smart Episode Filters
	SavePath                           *string      `json:"save_path"`                           // Default save path for torrents, separated by slashes
	SavePathChangedTmmEnabled          *bool        `json:"save_path_changed_tmm_enabled"`       // True if torrent should be relocated when the default save path changes
	SaveResumeDataInterval             *int         `json:"save_resume_data_interval"`           // Save resume data interval in min
	ScanDirs                           *interface{} `json:"scan_dirs"`                           // Property: directory to watch for torrent files, value: where torrents loaded from this directory should be downloaded to (see list of possible values below). Slashes are used as path separators; multiple key/value pairs can be specified
	ScheduleFromHour                   *int         `json:"schedule_from_hour"`                  // Scheduler starting hour
	ScheduleFromMin                    *int         `json:"schedule_from_min"`                   // Scheduler starting minute
	ScheduleToHour                     *int         `json:"schedule_to_hour"`                    // Scheduler ending hour
	ScheduleToMin                      *int         `json:"schedule_to_min"`                     // Scheduler ending minute
	SchedulerDays                      *int         `json:"scheduler_days"`                      // Scheduler days. See possible values here below [TODO]
	SchedulerEnabled                   *bool        `json:"scheduler_enabled"`                   // True if alternative limits should be applied according to schedule
	SendBufferLowWatermark             *int         `json:"send_buffer_low_watermark"`           // Send buffer low watermark in KiB
	SendBufferWatermark                *int         `json:"send_buffer_watermark"`               // Send buffer watermark in KiB
	SendBufferWatermarkFactor          *int         `json:"send_buffer_watermark_factor"`        // Send buffer watermark factor in percent
	SlowTorrentDlRateThreshold         *int         `json:"slow_torrent_dl_rate_threshold"`      // Download rate in KiB/s for a torrent to be considered "slow"
	SlowTorrentInactiveTimer           *int         `json:"slow_torrent_inactive_timer"`         // Seconds a torrent should be inactive before considered "slow"
	SlowTorrentUlRateThreshold         *int         `json:"slow_torrent_ul_rate_threshold"`      // Upload rate in KiB/s for a torrent to be considered "slow"
	SocketBacklogSize                  *int         `json:"socket_backlog_size"`                 // Socket backlog size
	StartPausedEnabled                 *bool        `json:"start_paused_enabled"`
	StopTrackerTimeout                 *int         `json:"stop_tracker_timeout"` // Timeout in seconds for a stopped announce request to trackers
	TempPath                           *string      `json:"temp_path"`            // Path for incomplete torrents, separated by slashes
	TempPathEnabled                    *bool        `json:"temp_path_enabled"`    // True if folder for incomplete torrents is enabled
	TorrentChangedTmmEnabled           *bool        `json:"torrent_changed_tmm_enabled"`
	UpLimit                            *int         `json:"up_limit"`                               // Global upload speed limit in KiB/s; -1 means no limit is applied
	UploadChokingAlgorithm             *int         `json:"upload_choking_algorithm"`               // Upload choking algorithm used (see list of possible values below)
	UploadSlotsBehavior                **int        `json:"upload_slots_behavior"`                  // Upload slots behavior used (see list of possible values below)
	UPnP                               *bool        `json:"upnp"`                                   // True if UPnP/NAT-PMP is enabled
	UseHTTPS                           *bool        `json:"use_https"`                              // True if WebUI HTTPS access is enabled
	UtpTCPMixedMode                    *int         `json:"utp_tcp_mixed_mode"`                     // μTP-TCP mixed mode algorithm (see list of possible values below)
	WebUIAddress                       *string      `json:"web_ui_address"`                         // IP address to use for the WebUI
	WebUIBanDuration                   *int         `json:"web_ui_ban_duration"`                    // WebUI access ban duration in seconds
	WebUIClickjackingProtectionEnabled *bool        `json:"web_ui_clickjacking_protection_enabled"` // True if WebUI clickjacking protection is enabled
	WebUICsrfProtectionEnabled         *bool        `json:"web_ui_csrf_protection_enabled"`         // True if WebUI CSRF protection is enabled
	WebUICustomHTTPHeaders             *string      `json:"web_ui_custom_http_headers"`             // List of custom http headers
	WebUIDomainList                    *string      `json:"web_ui_domain_list"`                     // Comma-separated list of domains to accept when performing Host header validation
	WebUIHostHeaderValidationEnabled   *bool        `json:"web_ui_host_header_validation_enabled"`  // True if WebUI host header validation is enabled
	WebUIHTTPSCertPath                 *string      `json:"web_ui_https_cert_path"`                 // Path to SSL certificate
	WebUIHTTPSKeyPath                  *string      `json:"web_ui_https_key_path"`                  // Path to SSL keyfile
	WebUIMaxAuthFailCount              *int         `json:"web_ui_max_auth_fail_count"`             // Maximum number of authentication failures before WebUI access ban
	WebUIPort                          *int         `json:"web_ui_port"`                            // WebUI port
	WebUISecureCookieEnabled           *bool        `json:"web_ui_secure_cookie_enabled"`           // True if WebUI cookie Secure flag is enabled
	WebUISessionTimeout                *int         `json:"web_ui_session_timeout"`                 // Seconds until WebUI is automatically signed off
	WebUIUPnP                          *bool        `json:"web_ui_upnp"`                            // True if UPnP is used for the WebUI port
	WebUIUseCustomHTTPHeadersEnabled   *bool        `json:"web_ui_use_custom_http_headers_enabled"` // Enable custom http headers
	WebUIUsername                      *string      `json:"web_ui_username"`                        // WebUI username
	WebUIPassword                      *string      `json:"web_ui_password"`                        // Plaintext WebUI password, not readable, write-only.
}

/*
upnp_lease_duration	integer	UPnP lease duration (0: Permanent lease)
*/
