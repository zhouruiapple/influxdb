// Utils
import {runQuery} from 'src/shared/apis/query'

// Types
import {LoadRowsOptions} from 'src/eventViewer/types'

export const runSyslogQuery = ({
  offset,
  limit,
  since,
  filter,
}: LoadRowsOptions) => {
  let filterFn = ''

  if (filter) {
    filterFn = `|> filter(fn: (r) => r.message =~ /${filter}/ or r.appname =~ /${filter}/ or r.facility =~ /${filter}/ or r.severity =~ /${filter}/)`
  }

  const query = `
from(bucket: "telegraf")
  |> range(start: -30d, stop: ${Math.round(since / 1000)})
  |> filter(fn: (r) => r._measurement == "syslog" and r._field == "message")
  |> group()
  |> keep(columns: ["_time", "_value", "appname", "facility", "host", "severity"])
  |> rename(columns: {"_time": "time", "_value": "message"})
  ${filterFn}
  |> limit(n: ${limit}, offset: ${offset})
`

  return runQuery('03f4f29fd0011000', query)
}
