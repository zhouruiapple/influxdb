// Libraries
import React, {FunctionComponent} from 'react'
import {connect} from 'react-redux'
import {LineLayerConfig} from '@influxdata/vis'
import {
  Form,
  Dropdown,
  MultiSelectDropdown,
  ComponentStatus,
} from '@influxdata/clockface'

// Utils
import {getNumericColumns, getGroupableColumns} from 'src/timeMachine/selectors'

// Types
import {AppState} from 'src/types'

interface StateProps {
  numericColumns: string[]
  groupColumns: string[]
}

interface OwnProps {
  layer: LineLayerConfig
  onSetLayer: (layer: LineLayerConfig) => void
}

type Props = OwnProps & StateProps

const VisLineLayerOptions: FunctionComponent<Props> = ({
  layer,
  onSetLayer,
  numericColumns,
  groupColumns,
}) => {
  const numericDropdownStatus = numericColumns.length
    ? ComponentStatus.Default
    : ComponentStatus.Disabled

  const groupDropdownStatus = groupColumns.length
    ? ComponentStatus.Default
    : ComponentStatus.Disabled

  console.log(layer.fill)

  return (
    <>
      <Form.Element label="X Column">
        <Dropdown
          selectedID={layer.x}
          onChange={x => onSetLayer({...layer, x})}
          status={numericDropdownStatus}
          titleText="None"
        >
          {numericColumns.map(columnName => (
            <Dropdown.Item id={columnName} key={columnName} value={columnName}>
              {columnName}
            </Dropdown.Item>
          ))}
        </Dropdown>
      </Form.Element>
      <Form.Element label="Y Column">
        <Dropdown
          selectedID={layer.y}
          onChange={y => onSetLayer({...layer, y})}
          status={numericDropdownStatus}
          titleText="None"
        >
          {numericColumns.map(columnName => (
            <Dropdown.Item id={columnName} key={columnName} value={columnName}>
              {columnName}
            </Dropdown.Item>
          ))}
        </Dropdown>
      </Form.Element>
      <Form.Element label="Group By">
        <MultiSelectDropdown
          selectedIDs={layer.fill}
          onChange={fill => onSetLayer({...layer, fill})}
          status={groupDropdownStatus}
        >
          {groupColumns.map(columnName => (
            <Dropdown.Item
              id={columnName}
              key={columnName}
              value={{id: columnName}}
            >
              {columnName}
            </Dropdown.Item>
          ))}
        </MultiSelectDropdown>
      </Form.Element>
    </>
  )
}

const mstp = (state: AppState) => {
  const numericColumns = getNumericColumns(state)
  const groupColumns = getGroupableColumns(state)

  return {numericColumns, groupColumns}
}

export default connect<StateProps>(mstp)(VisLineLayerOptions)
