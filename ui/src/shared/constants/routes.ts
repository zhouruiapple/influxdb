export const LOGIN = '/login'
export const LOGOUT = '/logout'
export const SIGNIN = '/signin'

export const BUCKETS = 'buckets'
export const BUCKET_ID = ':bucketID'

export const CLIENT_LIBS = 'client-libraries'

export const DASHBOARDS = 'dashboards'
export const DASHBOARD_ID = ':dashboardID'

export const ORGS = 'orgs'
export const ORG_ID = ':orgID'

export const SCRAPERS = 'scrapers'

export const SETTINGS = 'settings'

export const TELEGRAFS = 'telegrafs'

export const TOKENS = 'tokens'

export const VARIABLES = 'variables'

export const BUCKETS_ROUTE = '/orgs/:orgID/load-data/buckets'
export const TELEGRAFS_ROUTE = '/orgs/:orgID/load-data/telegrafs'
export const SCRAPERS_ROUTE = '/orgs/:orgID/load-data/scrapers'
export const TOKENS_ROUTE = '/orgs/:orgID/load-data/tokens'
export const CLIENT_LIBRARIES_ROUTE = '/orgs/:orgID/load-data/client-libraries'

export const EXPLORER_ROUTE = '/orgs/:orgID/data-explorer'
export const DASHBOARDS_ROUTE = '/orgs/:orgID/dashboards-list'
export const TASKS_ROUTE = '/orgs/:orgID/tasks'
export const ALERTING_ROUTE = '/orgs/:orgID/alerting'

export const VARIABLES_ROUTE = '/orgs/:orgID/settings/variables'
export const TEMPLATES_ROUTE = '/orgs/:orgID/settings/templates'
export const LABELS_ROUTE = '/orgs/:orgID/settings/labels'
export const COMMUNITY_TEMPLATES_IMPORT_ROUTE = `${TEMPLATES_ROUTE}/import/:directory/:templateName/:templateExtension`
