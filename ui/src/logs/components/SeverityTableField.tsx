// Libraries
import React, {FC} from 'react'

interface Props {
  row: {severity: string}
}

const SeverityTableField: FC<Props> = ({row: {severity}}) => {
  return (
    <div
      className={`severity-table-field severity-table-field--${severity.toLowerCase()}`}
    >
      {severity}
    </div>
  )
}

export default SeverityTableField
