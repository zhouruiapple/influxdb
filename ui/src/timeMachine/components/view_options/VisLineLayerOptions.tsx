// Libraries
import React, {FunctionComponent} from 'react'
import {connect} from 'react-redux'
import {LineLayerConfig} from '@influxdata/vis'
import {
  Form,
  Dropdown,
  // MultiSelectDropdown,
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
  // groupColumns,
}) => {
  const numericDropdownStatus = numericColumns.length
    ? ComponentStatus.Default
    : ComponentStatus.Disabled

  // const groupDropdownStatus = groupColumns.length
  //   ? ComponentStatus.Default
  //   : ComponentStatus.Disabled

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
      <Form.Element label="Interpolation">
        <Dropdown
          selectedID={layer.interpolation}
          onChange={interpolation => onSetLayer({...layer, interpolation})}
        >
          <Dropdown.Item id="linear" key="linear" value="linear">
            Linear
          </Dropdown.Item>
          <Dropdown.Item id="monotoneX" key="monotoneX" value="monotoneX">
            Smooth (X-Monotonic)
          </Dropdown.Item>
          <Dropdown.Item id="monotoneY" key="monotoneY" value="monotoneY">
            Smooth (Y-Monotonic)
          </Dropdown.Item>
          <Dropdown.Item id="step" key="step" value="step">
            Step
          </Dropdown.Item>
          <Dropdown.Item id="stepBefore" key="stepBefore" value="stepBefore">
            Step Before
          </Dropdown.Item>
          <Dropdown.Item id="stepAfter" key="stepAfter" value="stepAfter">
            Step After
          </Dropdown.Item>
        </Dropdown>
      </Form.Element>
      {/*
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
      */}
    </>
  )
}

const mstp = (state: AppState) => {
  const numericColumns = getNumericColumns(state)
  const groupColumns = getGroupableColumns(state)

  return {numericColumns, groupColumns}
}

export default connect<StateProps>(mstp)(VisLineLayerOptions)
