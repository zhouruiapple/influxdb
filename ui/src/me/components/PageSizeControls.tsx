// Libraries
import React, {PureComponent} from 'react'
import {connect} from 'react-redux'

// Components
import {
  Panel,
  ComponentSize,
  Radio,
  ButtonShape,
  FlexBox,
  FlexDirection,
  JustifyContent,
} from '@influxdata/clockface'

// Types
import {AppState} from 'src/types'
import {togglePageSize} from 'src/userSettings/actions/index'

interface StateProps {
  pageSize: ComponentSize
}

interface DispatchProps {
  onTogglePageSize: typeof togglePageSize
}

type Props = StateProps & DispatchProps

class PageSizeControls extends PureComponent<Props> {
  public render() {
    const {pageSize} = this.props

    return (
      <Panel>
        <Panel.Header>
          <Panel.Title>Options</Panel.Title>
        </Panel.Header>
        <Panel.Body>
          <FlexBox
            direction={FlexDirection.Row}
            justifyContent={JustifyContent.SpaceBetween}
          >
            <p>Page Size</p>
            <Radio shape={ButtonShape.Square} className="page-size-controls">
              <Radio.Button
                titleText="Change page size to extra-small"
                value={ComponentSize.ExtraSmall}
                id="xs"
                active={pageSize === ComponentSize.ExtraSmall}
                onClick={this.handleRadioClick}
              >
                <div className="page-size-controls--icon__xs" />
              </Radio.Button>
              <Radio.Button
                titleText="Change page size to small"
                value={ComponentSize.Small}
                id="sm"
                active={pageSize === ComponentSize.Small}
                onClick={this.handleRadioClick}
              >
                <div className="page-size-controls--icon__sm" />
              </Radio.Button>
              <Radio.Button
                titleText="Change page size to medium"
                value={ComponentSize.Medium}
                id="md"
                active={pageSize === ComponentSize.Medium}
                onClick={this.handleRadioClick}
              >
                <div className="page-size-controls--icon__md" />
              </Radio.Button>
              <Radio.Button
                titleText="Change page size to large"
                value={ComponentSize.Large}
                id="lg"
                active={pageSize === ComponentSize.Large}
                onClick={this.handleRadioClick}
              >
                <div className="page-size-controls--icon__lg" />
              </Radio.Button>
            </Radio>
          </FlexBox>
        </Panel.Body>
      </Panel>
    )
  }

  private handleRadioClick = (pageSize: ComponentSize): void => {
    const {onTogglePageSize} = this.props

    onTogglePageSize(pageSize)
  }
}

const mstp = (state: AppState) => {
  const {
    userSettings: {pageSize},
  } = state

  return {pageSize}
}

const mdtp = {
  onTogglePageSize: togglePageSize,
}

export default connect<StateProps, DispatchProps>(
  mstp,
  mdtp
)(PageSizeControls)
