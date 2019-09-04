// Libraries
import React, {SFC} from 'react'

// Components
import CellHeaderNote from 'src/shared/components/cells/CellHeaderNote'

interface Props {
  name: string
  note: string
}

const CellHeader: SFC<Props> = ({name, note}) => (
  <div className="cell--header">
    <div className="cell--drag-handle">
      <div className="cell--drag-icon" />
    </div>
    <div className="cell--name">
      {name}
      {note && <CellHeaderNote note={note} />}
    </div>
  </div>
)

export default CellHeader
