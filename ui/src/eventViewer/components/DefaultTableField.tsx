// Libraries
import React, {FC} from 'react'

// Types
import {Row} from 'src/eventViewer/types'

interface Props {
  row: Row
  field: string
}

const DefaultTableField: FC<Props> = ({row, field}) => {
  return <div className="default-table-field">{String(row[field])}</div>
}

export default DefaultTableField
