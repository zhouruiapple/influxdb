// Libraries
import React, {PureComponent, ReactNode} from 'react'

// Components
import {DapperScrollbars} from '@influxdata/clockface'

interface Props {
  scrollable: boolean
  addPadding: boolean
  testID: string
}

export default class BuilderCardBody extends PureComponent<Props> {
  public static defaultProps = {
    scrollable: true,
    addPadding: true,
    testID: 'builder-card--body',
  }

  public render() {
    const {scrollable, testID} = this.props

    if (scrollable) {
      return (
        <div className="builder-card--body" data-testid={testID}>
          <DapperScrollbars
            autoSize={true}
            style={{maxWidth: '100%', maxHeight: '100%'}}
          >
            {this.children}
          </DapperScrollbars>
        </div>
      )
    }

    return (
      <div className="builder-card--body" data-testid={testID}>
        {this.children}
      </div>
    )
  }

  private get children(): JSX.Element | ReactNode {
    const {addPadding, children} = this.props

    if (addPadding) {
      return <div className="builder-card--contents">{children}</div>
    }

    return children
  }
}
