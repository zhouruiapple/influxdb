// Libraries
import React, {/* useState, */ FC} from 'react'
import {range} from 'lodash'
import {Page} from '@influxdata/clockface'
import {fromFlux} from '@influxdata/giraffe'

// Components
import EventViewer from 'src/eventViewer/components/EventViewer'
import EventTable from 'src/eventViewer/components/EventTable'
import BackToTopButton from 'src/eventViewer/components/BackToTopButton'
import LimitDropdown from 'src/eventViewer/components/LimitDropdown'
import LogsSearchBar from 'src/logs/components/LogsSearchBar'
import SeverityTableField from 'src/logs/components/SeverityTableField'
import TimeTableField from 'src/logs/components/TimeTableField'
// import LogsHistogram from 'src/logs/components/LogsHistogram'

// Utils
import {runSyslogQuery} from 'src/logs/utils/loadSyslogRows'

// Types
import {FieldComponents, LoadRows, Row} from 'src/eventViewer/types'

const FIELD_COMPONENTS: FieldComponents = {
  time: TimeTableField,
  severity: SeverityTableField,
}

const FIELD_WIDTHS = {
  time: 160,
  message: 300,
  facility: 80,
  appname: 80,
  severity: 50,
  host: 100,
}

const LogsIndex: FC = () => {
  // const [table, setTable] = useState(null)

  const loadRows: LoadRows = options => {
    const {promise: queryPromise, cancel} = runSyslogQuery(options)

    const rowPromise = queryPromise.then<Row[]>(resp => {
      if (resp.type !== 'SUCCESS') {
        return Promise.reject(new Error(resp.message))
      }

      const {table: nextTable} = fromFlux(resp.csv)

      // setTable(nextTable)

      const rows = range(nextTable.length).map(i => {
        const row = {}

        for (const key of nextTable.columnKeys) {
          row[key] = nextTable.getColumn(key)[i]
        }

        return row
      })

      return rows
    })

    return {promise: rowPromise, cancel}
  }

  return (
    <EventViewer loadRows={loadRows}>
      {props => (
        <Page titleTag="Logs | InfluxDB 2.0" className="logs-index">
          <Page.Header fullWidth={true}>
            <div className="logs-index--header">
              <Page.Title title="Logs" />
              <div className="logs-index--controls">
                <BackToTopButton {...props} />
                <LimitDropdown {...props} />
                <LogsSearchBar {...props} />
              </div>
            </div>
          </Page.Header>
          <Page.Contents
            fullWidth={true}
            fullHeight={true}
            scrollable={false}
            className="logs-index--contents"
          >
            <div className="logs-index--table">
              <EventTable
                {...props}
                fields={[
                  'time',
                  'severity',
                  'facility',
                  'appname',
                  'message',
                  'host',
                ]}
                fieldWidths={FIELD_WIDTHS}
                fieldComponents={FIELD_COMPONENTS}
              />
              {/* <LogsHistogram table={table} /> */}
            </div>
          </Page.Contents>
        </Page>
      )}
    </EventViewer>
  )
}

export default LogsIndex
