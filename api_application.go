package qbtapi

import (
	"context"
	"fmt"
)

/*
	Application
	https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#application
*/

const (
	applicationAPIName = "app"
)

// GetApplicationVersion returns the application version.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-application-version
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
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-api-version
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
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-build-info
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
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#shutdown-application
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
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-application-preferences
type ApplicationPreferences struct {
	Locale                             *string        `json:"locale,omitempty"`                                 // True if a subfolder should be created when adding a torrent
	CreateSubfolderEnabled             *bool          `json:"create_subfolder_enabled,omitempty"`               // True if a subfolder should be created when adding a torrent
	StartPausedEnabled                 *bool          `json:"start_paused_enabled,omitempty"`                   // True if torrents should be added in a Paused state
	AutoDeleteMode                     *int           `json:"auto_delete_mode,omitempty"`                       // TODO
	PreallocateAll                     *bool          `json:"preallocate_all,omitempty"`                        // True if disk space should be pre-allocated for all files
	IncompleteFilesExt                 *bool          `json:"incomplete_files_ext,omitempty"`                   // True if ".!qB" should be appended to incomplete files
	AutoTMMEnabled                     *bool          `json:"auto_tmm_enabled,omitempty"`                       // True if Automatic Torrent Management is enabled by default
	TorrentChangedTmmEnabled           *bool          `json:"torrent_changed_tmm_enabled,omitempty"`            // True if torrent should be relocated when its Category changes
	SavePathChangedTmmEnabled          *bool          `json:"save_path_changed_tmm_enabled,omitempty"`          // True if torrent should be relocated when the default save path changes
	CategoryChangedTmmEnabled          *bool          `json:"category_changed_tmm_enabled,omitempty"`           // True if torrent should be relocated when its Category's save path changes
	SavePath                           *string        `json:"save_path,omitempty"`                              // Default save path for torrents, separated by slashes
	TempPathEnabled                    *bool          `json:"temp_path_enabled,omitempty"`                      // True if folder for incomplete torrents is enabled
	TempPath                           *string        `json:"temp_path,omitempty"`                              // Path for incomplete torrents, separated by slashes
	ScanDirs                           map[string]any `json:"scan_dirs,omitempty"`                              // Directories to watch for torrent files, value: where torrents loaded from this directory should be downloaded to (see list of possible values below). Slashes are used as path separators.
	ExportDir                          *string        `json:"export_dir,omitempty"`                             // Path to directory to copy .torrent files to. Slashes are used as path separators
	ExportDirFin                       *string        `json:"export_dir_fin,omitempty"`                         // Path to directory to copy .torrent files of completed downloads to. Slashes are used as path separators
	MailNotificationEnabled            *bool          `json:"mail_notification_enabled,omitempty"`              // True if e-mail notification should be enabled
	MailNotificationSender             *string        `json:"mail_notification_sender,omitempty"`               // e-mail where notifications should originate from
	MailNotificationEmail              *string        `json:"mail_notification_email,omitempty"`                // e-mail to send notifications to
	MailNotificationSMTP               *string        `json:"mail_notification_smtp,omitempty"`                 // smtp server for e-mail notifications
	MailNotificationSslEnabled         *bool          `json:"mail_notification_ssl_enabled,omitempty"`          // True if smtp server requires SSL connection
	MailNotificationAuthEnabled        *bool          `json:"mail_notification_auth_enabled,omitempty"`         // True if smtp server requires authentication
	MailNotificationUsername           *string        `json:"mail_notification_username,omitempty"`             // Username for smtp authentication
	MailNotificationPassword           *string        `json:"mail_notification_password,omitempty"`             // Password for smtp authentication
	AutorunEnabled                     *bool          `json:"autorun_enabled,omitempty"`                        // True if external program should be run after torrent has finished downloading
	AutorunProgram                     *string        `json:"autorun_program,omitempty"`                        // Program path/name/arguments to run if autorun_enabled is enabled; path is separated by slashes; you can use %f and %n arguments, which will be expanded by qBittorent as path_to_torrent_file and torrent_name (from the GUI; not the .torrent file name) respectively
	QueueingEnabled                    *bool          `json:"queueing_enabled,omitempty"`                       // True if torrent queuing is enabled
	MaxActiveDownloads                 *int           `json:"max_active_downloads,omitempty"`                   // Maximum number of active simultaneous downloads
	MaxActiveTorrents                  *int           `json:"max_active_torrents,omitempty"`                    // Maximum number of active simultaneous downloads and uploads
	MaxActiveUploads                   *int           `json:"max_active_uploads,omitempty"`                     // Maximum number of active simultaneous uploads
	DontCountSlowTorrents              *bool          `json:"dont_count_slow_torrents,omitempty"`               // If true torrents w/o any activity (stalled ones) will not be counted towards max_active_* limits; see https://www.libtorrent.org/reference-Settings.html#dont_count_slow_torrents for more information
	SlowTorrentDlRateThreshold         *int           `json:"slow_torrent_dl_rate_threshold,omitempty"`         // Download rate in KiB/s for a torrent to be considered "slow"
	SlowTorrentUlRateThreshold         *int           `json:"slow_torrent_ul_rate_threshold,omitempty"`         // Upload rate in KiB/s for a torrent to be considered "slow"
	SlowTorrentInactiveTimer           *int           `json:"slow_torrent_inactive_timer,omitempty"`            // Seconds a torrent should be inactive before considered "slow"
	MaxRatioEnabled                    *bool          `json:"max_ratio_enabled,omitempty"`                      // True if share ratio limit is enabled
	MaxRatio                           *int           `json:"max_ratio,omitempty"`                              // Get the global share ratio limit
	MaxRatioAct                        *int           `json:"max_ratio_act,omitempty"`                          // Action performed when a torrent reaches the maximum share ratio.
	ListenPort                         *int           `json:"listen_port,omitempty"`                            // Port for incoming connections
	UPnP                               *bool          `json:"upnp,omitempty"`                                   // True if UPnP/NAT-PMP is enabled
	RandomPort                         *bool          `json:"random_port,omitempty"`                            // True if the port is randomly selected
	DlLimit                            *int           `json:"dl_limit,omitempty"`                               // Global download speed limit in KiB/s; -1 means no limit is applied
	UpLimit                            *int           `json:"up_limit,omitempty"`                               // Global upload speed limit in KiB/s; -1 means no limit is applied
	MaxConnec                          *int           `json:"max_connec,omitempty"`                             // Maximum global number of simultaneous connections
	MaxConnecPerTorrent                *int           `json:"max_connec_per_torrent,omitempty"`                 // Maximum number of simultaneous connections per torrent
	MaxUploads                         *int           `json:"max_uploads,omitempty"`                            // Maximum number of upload slots
	MaxUploadsPerTorrent               *int           `json:"max_uploads_per_torrent,omitempty"`                // Maximum number of upload slots per torrent
	StopTrackerTimeout                 *int           `json:"stop_tracker_timeout,omitempty"`                   // Timeout in seconds for a stopped announce request to trackers
	EnablePieceExtentAffinity          *bool          `json:"enable_piece_extent_affinity,omitempty"`           // True if the advanced libtorrent option https://www.libtorrent.org/reference-Settings.html#piece_extent_affinity is enabled
	BittorrentProtocol                 *int           `json:"bittorrent_protocol,omitempty"`                    // Bittorrent Protocol to use
	LimitUtpRate                       *bool          `json:"limit_utp_rate,omitempty"`                         // True if [du]l_limit should be applied to uTP connections; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	LimitTCPOverhead                   *bool          `json:"limit_tcp_overhead,omitempty"`                     // True if [du]l_limit should be applied to estimated TCP overhead (service data: e.g. packet headers)
	LimitLanPeers                      *bool          `json:"limit_lan_peers,omitempty"`                        // True if [du]l_limit should be applied to peers on the LAN
	AltDlLimit                         *int           `json:"alt_dl_limit,omitempty"`                           // Alternative global download speed limit in KiB/s
	AltUpLimit                         *int           `json:"alt_up_limit,omitempty"`                           // Alternative global upload speed limit in KiB/s
	SchedulerEnabled                   *bool          `json:"scheduler_enabled,omitempty"`                      // True if alternative limits should be applied according to schedule
	ScheduleFromHour                   *int           `json:"schedule_from_hour,omitempty"`                     // Scheduler starting hour
	ScheduleFromMin                    *int           `json:"schedule_from_min,omitempty"`                      // Scheduler starting minute
	ScheduleToHour                     *int           `json:"schedule_to_hour,omitempty"`                       // Scheduler ending hour
	ScheduleToMin                      *int           `json:"schedule_to_min,omitempty"`                        // Scheduler ending minute
	SchedulerDays                      *int           `json:"scheduler_days,omitempty"`                         // Scheduler days. See possible values here below
	DHT                                *bool          `json:"dht,omitempty"`                                    // True if DHT is enabled
	PeX                                *bool          `json:"pex,omitempty"`                                    // True if PeX is enabled
	LSD                                *bool          `json:"lsd,omitempty"`                                    // True if LSD is enabled
	Encryption                         *int           `json:"encryption,omitempty"`                             // Transmission encryption usage
	AnonymousMode                      *bool          `json:"anonymous_mode,omitempty"`                         // If true anonymous mode will be enabled; read more at https://github.com/qbittorrent/qBittorrent/wiki/Anonymous-Mode; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	ProxyType                          *int           `json:"proxy_type,omitempty"`                             // Proxy usage
	ProxyIP                            *string        `json:"proxy_ip,omitempty"`                               // Proxy IP address or domain name
	ProxyPort                          *int           `json:"proxy_port,omitempty"`                             // Proxy port
	ProxyPeerConnections               *bool          `json:"proxy_peer_connections,omitempty"`                 // True if peer and web seed connections should be proxified; this option will have any effect only in qBittorent built against libtorrent version 0.16.X and higher
	ProxyAuthEnabled                   *bool          `json:"proxy_auth_enabled,omitempty"`                     // True proxy requires authentication; doesn't apply to SOCKS4 proxies
	ProxyUsername                      *string        `json:"proxy_username,omitempty"`                         // Username for proxy authentication
	ProxyPassword                      *string        `json:"proxy_password,omitempty"`                         // Password for proxy authentication
	ProxyTorrentsOnly                  *bool          `json:"proxy_torrents_only,omitempty"`                    // True if proxy is only used for torrents
	IPFilterEnabled                    *bool          `json:"ip_filter_enabled,omitempty"`                      // True if external IP filter should be enabled
	IPFilterPath                       *string        `json:"ip_filter_path,omitempty"`                         // Path to IP filter file (.dat, .p2p, .p2b files are supported); path is separated by slashes
	IPFilterTrackers                   *bool          `json:"ip_filter_trackers,omitempty"`                     // True if IP filters are applied to trackers
	WebUIDomainList                    *string        `json:"web_ui_domain_list,omitempty"`                     // Semicolon-separated list of domains to accept when performing Host header validation
	WebUIAddress                       *string        `json:"web_ui_address,omitempty"`                         // IP address to use for the WebUI
	WebUIPort                          *int           `json:"web_ui_port,omitempty"`                            // WebUI
	WebUIUpnp                          *bool          `json:"web_ui_upnp,omitempty"`                            // True if UPnP is used for the WebUI port
	WebUIUsername                      *string        `json:"web_ui_username,omitempty"`                        // WebUI username
	WebUIPassword                      *string        `json:"web_ui_password,omitempty"`                        // For API ≥ v2.3.0: Plaintext WebUI password, not readable, write-only. For API < v2.3.0: MD5 hash of WebUI password, hash is generated from the following string: username:Web UI Access:plain_text_web_ui_password
	WebUICsrfProtectionEnabled         *bool          `json:"web_ui_csrf_protection_enabled,omitempty"`         // True if WebUI CSRF protection is enabled
	WebUIClickjackingProtectionEnabled *bool          `json:"web_ui_clickjacking_protection_enabled,omitempty"` // True if WebUI clickjacking protection is enabled
	WebUISecureCookieEnabled           *bool          `json:"web_ui_secure_cookie_enabled,omitempty"`           // True if WebUI cookie Secure flag is enabled
	WebUIMaxAuthFailCount              *int           `json:"web_ui_max_auth_fail_count,omitempty"`             // Maximum number of authentication failures before WebUI access ban
	WebUIBanDuration                   *int           `json:"web_ui_ban_duration,omitempty"`                    // WebUI access ban duration in seconds
	WebUISessionTimeout                *int           `json:"web_ui_session_timeout,omitempty"`                 // 	Seconds until WebUI is automatically signed off
	WebUIHostHeaderValidationEnabled   *bool          `json:"web_ui_host_header_validation_enabled,omitempty"`  // True if WebUI host header validation is enabled
	BypassLocalAuth                    *bool          `json:"bypass_local_auth,omitempty"`                      // True if authentication challenge for loopback address (127.0.0.1) should be disabled
	BypassAuthSubnetWhitelistEnabled   *bool          `json:"bypass_auth_subnet_whitelist_enabled,omitempty"`   // True if webui authentication should be bypassed for clients whose ip resides within (at least) one of the subnets on the whitelist
	BypassAuthSubnetWhitelist          *string        `json:"bypass_auth_subnet_whitelist,omitempty"`           // (White)list of ipv4/ipv6 subnets for which webui authentication should be bypassed; list entries are separated by commas
	AlternativeWebuiEnabled            *bool          `json:"alternative_webui_enabled,omitempty"`              // True if an alternative WebUI should be used
	AlternativeWebuiPath               *string        `json:"alternative_webui_path,omitempty"`                 // File path to the alternative WebUI
	UseHTTPS                           *bool          `json:"use_https,omitempty"`                              // True if WebUI HTTPS access is enabled
	WebUIHTTPSKeyPath                  *string        `json:"web_ui_https_key_path,omitempty"`                  // Path to SSL keyfile
	WebUIHTTPSCertPath                 *string        `json:"web_ui_https_cert_path,omitempty"`                 // Path to SSL certificate
	DynDNSEnabled                      *bool          `json:"dyndns_enabled,omitempty"`                         // True if server DNS should be updated dynamically
	DynDNSService                      *int           `json:"dyndns_service,omitempty"`                         // DynDNS service to use
	DynDNSUsername                     *string        `json:"dyndns_username,omitempty"`                        // Username for DDNS service
	DynDNSPassword                     *string        `json:"dyndns_password,omitempty"`                        // Password for DDNS service
	DynDNSDomain                       *string        `json:"dyndns_domain,omitempty"`                          // Your DDNS domain name
	RSSRefreshInterval                 *int           `json:"rss_refresh_interval,omitempty"`                   // RSS refresh interval
	RSSMaxArticlesPerFeed              *int           `json:"rss_max_articles_per_feed,omitempty"`              // Max stored articles per RSS feed
	RSSProcessingEnabled               *bool          `json:"rss_processing_enabled,omitempty"`                 // Enable processing of RSS feeds
	RSSAutoDownloadingEnabled          *bool          `json:"rss_auto_downloading_enabled,omitempty"`           // Enable auto-downloading of torrents from the RSS feeds
	RSSDownloadRepackProperEpisodes    *bool          `json:"rss_download_repack_proper_episodes,omitempty"`    // Enable downloading of repack/proper Episodes
	RSSSmartEpisodeFilters             *string        `json:"rss_smart_episode_filters,omitempty"`              // List of RSS Smart Episode Filters
	AddTrackersEnabled                 *bool          `json:"add_trackers_enabled,omitempty"`                   // Enable automatic adding of trackers to new torrents
	AddTrackers                        *string        `json:"add_trackers,omitempty"`                           // List of trackers to add to new torrent
	WebUIUseCustomHTTPHeadersEnabled   *bool          `json:"web_ui_use_custom_http_headers_enabled,omitempty"` // Enable custom http headers
	WebUICustomHTTPHeaders             *string        `json:"web_ui_custom_http_headers,omitempty"`             // List of custom http headers
	MaxSeedingTimeEnabled              *bool          `json:"max_seeding_time_enabled,omitempty"`               // True enables max seeding time
	MaxSeedingTime                     *int           `json:"max_seeding_time,omitempty"`                       // Number of minutes to seed a torrent
	AnnounceIP                         *string        `json:"announce_ip,omitempty"`                            // TODO
	AnnounceToAllTiers                 *bool          `json:"announce_to_all_tiers,omitempty"`                  // True always announce to all tiers
	AnnounceToAllTrackers              *bool          `json:"announce_to_all_trackers,omitempty"`               // True always announce to all trackers in a tier
	AsyncIoThreads                     *int           `json:"async_io_threads,omitempty"`                       // Number of asynchronous I/O threads
	BannedIPs                          *string        `json:"banned_IPs,omitempty"`                             // List of banned IPs
	CheckingMemoryUse                  *int           `json:"checking_memory_use,omitempty"`                    // Outstanding memory when checking torrents in MiB
	CurrentInterfaceAddress            *string        `json:"current_interface_address,omitempty"`              // IP Address to bind to. Empty String means All addresses
	CurrentNetworkInterface            *string        `json:"current_network_interface,omitempty"`              // Network Interface used
	DiskCache                          *int           `json:"disk_cache,omitempty"`                             // Disk cache used in MiB
	DiskCacheTTL                       *int           `json:"disk_cache_ttl,omitempty"`                         // Disk cache expiry interval in seconds
	EmbeddedTrackerPort                *int           `json:"embedded_tracker_port,omitempty"`                  // Port used for embedded tracker
	EnableCoalesceReadWrite            *bool          `json:"enable_coalesce_read_write,omitempty"`             // True enables coalesce reads & writes
	EnableEmbeddedTracker              *bool          `json:"enable_embedded_tracker,omitempty"`                // True enables embedded tracker
	EnableMultiConnectionsFromSameIP   *bool          `json:"enable_multi_connections_from_same_ip,omitempty"`  // True allows multiple connections from the same IP address
	EnableOSCache                      *bool          `json:"enable_os_cache,omitempty"`                        // True enables OS cache
	EnableUploadSuggestions            *bool          `json:"enable_upload_suggestions,omitempty"`              // True enables sending of upload piece suggestions
	FilePoolSize                       *int           `json:"file_pool_size,omitempty"`                         // File pool size
	OutgoingPortsMax                   *int           `json:"outgoing_ports_max,omitempty"`                     // Maximal outgoing port (0: Disabled)
	OutgoingPortsMin                   *int           `json:"outgoing_ports_min,omitempty"`                     // Minimal outgoing port (0: Disabled)
	RecheckCompletedTorrents           *bool          `json:"recheck_completed_torrents,omitempty"`             // True rechecks torrents on completion
	ResolvePeerCountries               *bool          `json:"resolve_peer_countries,omitempty"`                 // True resolves peer countries
	SaveResumeDataInterval             *int           `json:"save_resume_data_interval,omitempty"`              // Save resume data interval in min
	SendBufferLowWatermark             *int           `json:"send_buffer_low_watermark,omitempty"`              // Send buffer low watermark in KiB
	SendBufferWatermark                *int           `json:"send_buffer_watermark,omitempty"`                  // Send buffer watermark in KiB
	SendBufferWatermarkFactor          *int           `json:"send_buffer_watermark_factor,omitempty"`           // Send buffer watermark factor in percent
	SocketBacklogSize                  *int           `json:"socket_backlog_size,omitempty"`                    // Socket backlog size
	UploadChokingAlgorithm             *int           `json:"upload_choking_algorithm,omitempty"`               // Upload choking algorithm used (see list of possible values below)
	UploadSlotsBehavior                *int           `json:"upload_slots_behavior,omitempty"`                  // Upload slots behavior used (see list of possible values below)
	UPnPLeaseDuration                  *int           `json:"upnp_lease_duration,omitempty"`                    // UPnP lease duration (0: Permanent lease)
	UTPTCPMixedMode                    *int           `json:"utp_tcp_mixed_mode,omitempty"`                     // μTP-TCP mixed mode algorithm (see list of possible values below)
}
