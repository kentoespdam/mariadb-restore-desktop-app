// Wails runtime + Go bindings mock for browser dev mode.
// Loaded ONLY when window.runtime is missing (i.e. not inside Wails).
(function () {
  if (window.runtime) return; // already injected by Wails

  const noop = () => {};
  const noopFn = () => Promise.resolve();

  // --- runtime mock ---
  window.runtime = {
    LogPrint: noop, LogTrace: noop, LogDebug: noop, LogInfo: noop,
    LogWarning: noop, LogError: noop, LogFatal: noop,
    EventsOnMultiple: () => noop,
    EventsOn: () => noop,
    EventsOff: noop, EventsOffAll: noop, EventsOnce: noop, EventsEmit: noop,
    WindowReload: noop, WindowReloadApp: noop,
    WindowSetAlwaysOnTop: noop, WindowSetSystemDefaultTheme: noop,
    WindowSetLightTheme: noop, WindowSetDarkTheme: noop,
    WindowCenter: noop, WindowSetTitle: noop,
    WindowFullscreen: noop, WindowUnfullscreen: noop,
    WindowIsFullscreen: () => false,
    WindowGetSize: () => ({ w: 1024, h: 768 }),
    WindowSetSize: noop, WindowSetMaxSize: noop, WindowSetMinSize: noop,
    WindowSetPosition: noop, WindowGetPosition: () => ({ x: 0, y: 0 }),
    WindowHide: noop, WindowShow: noop,
    WindowMaximise: noop, WindowToggleMaximise: noop, WindowUnmaximise: noop,
    WindowIsMaximised: () => false,
    WindowMinimise: noop, WindowUnminimise: noop,
    WindowSetBackgroundColour: noop,
    ScreenGetAll: () => [],
    WindowIsMinimised: () => false, WindowIsNormal: () => true,
    BrowserOpenURL: noop, Environment: () => ({ buildType: 'dev' }),
    Quit: noop, Hide: noop, Show: noop,
    ClipboardGetText: () => '', ClipboardSetText: noop,
    OnFileDrop: noop, OnFileDropOff: noop,
    CanResolveFilePaths: () => true, ResolveFilePaths: (files) => {
      for (const f of files) {
        Object.defineProperty(f, 'path', { value: f.name, configurable: true });
      }
    },
    InitializeNotifications: noop, CleanupNotifications: noop,
    IsNotificationAvailable: () => false,
    RequestNotificationAuthorization: noop, CheckNotificationAuthorization: () => false,
    SendNotification: noop, SendNotificationWithActions: noop,
    RegisterNotificationCategory: noop, RemoveNotificationCategory: noop,
    RemoveAllPendingNotifications: noop, RemovePendingNotification: noop,
    RemoveAllDeliveredNotifications: noop, RemoveDeliveredNotification: noop,
    RemoveNotification: noop,
  };

  // --- Go bindings mock ---
  window.go = window.go || {};
  window.go.app = window.go.app || {};
  window.go.app.App = {
    // Profile
    ListServerProfiles: () => Promise.resolve([
      { id: 'mock-1', name: 'docker-root', host: '127.0.0.1', port: 3307, user: 'root', hasPassword: true, sslMode: 'preferred' },
    ]),
    CreateServerProfile: (input) => Promise.resolve('mock-' + Date.now()),
    UpdateServerProfile: noopFn,
    DeleteServerProfile: noopFn,
    // Backup
    StartBackup: (req) => Promise.resolve('backup-job-' + Date.now()),
    CancelBackup: noopFn,
    // Restore
    OpenDumpFileDialog: () => Promise.resolve('/tmp/mock-dump.sql'),
    AnalyzeDump: (path) => Promise.resolve(5),
    ListCatalogObjects: (path) => Promise.resolve([
      { id: 1, database: 'shop', name: 'products', type: 'CREATE_TABLE', startByte: 0, endByte: 500 },
      { id: 2, database: 'shop', name: 'products', type: 'INSERT', startByte: 501, endByte: 1200 },
      { id: 3, database: 'shop', name: 'orders', type: 'CREATE_TABLE', startByte: 1201, endByte: 1800 },
      { id: 4, database: 'shop', name: 'orders', type: 'INSERT', startByte: 1801, endByte: 2500 },
      { id: 5, database: 'testdb', name: 'users', type: 'CREATE_TABLE', startByte: 2501, endByte: 3000 },
    ]),
    StartFullRestore: (req) => Promise.resolve('restore-job-' + Date.now()),
    StartPartialRestore: (req) => Promise.resolve('restore-job-' + Date.now()),
    CancelRestore: noopFn,
    // Settings
    GetSettings: () => Promise.resolve({
      exeDir: '/tmp/liveapp',
      catalogPath: '/tmp/liveapp/catalog.sqlite',
      appKeyPath: '/tmp/liveapp/app.key',
      mariadbPath: '/usr/bin/mariadb',
      mariadbDumpPath: '/usr/bin/mariadb-dump',
      keyBits: 256,
    }),
    SaveSettings: noopFn,
    ResetAndReinit: () => Promise.resolve({ triggered: 'reset' }),
    // Recovery
    HandleMissingKey: () => Promise.resolve('missing_key'),
    RecoveryDecision: noopFn,
    // App
    RebindCtx: noopFn,
    Close: noopFn,
  };
})();
