// Libraries
import React, {PureComponent, CSSProperties, RefObject} from 'react'
import classnames from 'classnames'

// Components
import BuilderCardHeader from 'src/timeMachine/components/builderCard/BuilderCardHeader'
import BuilderCardMenu from 'src/timeMachine/components/builderCard/BuilderCardMenu'
import BuilderCardBody from 'src/timeMachine/components/builderCard/BuilderCardBody'
import BuilderCardEmpty from 'src/timeMachine/components/builderCard/BuilderCardEmpty'
import Handle from 'src/timeMachine/components/builderCard/ResizableBuilderCardHandle'

// Constants
import {
  BUILDER_CARD_DEFAULT_WIDTH,
  BUILDER_CARD_MIN_WIDTH,
  BUILDER_CARD_MAX_WIDTH,
} from 'src/timeMachine/constants'

interface Props {
  testID: string
  className?: string
}

interface State {
  widthPixels: number
}

export default class ResizableBuilderCard extends PureComponent<Props, State> {
  public static Header = BuilderCardHeader
  public static Menu = BuilderCardMenu
  public static Body = BuilderCardBody
  public static Empty = BuilderCardEmpty

  public state: State = {
    widthPixels: BUILDER_CARD_DEFAULT_WIDTH,
  }

  public static defaultProps = {
    testID: 'builder-card',
  }

  public cardRef: RefObject<HTMLDivElement> = React.createRef()

  public render() {
    const {widthPixels} = this.state
    const {children, testID, className} = this.props

    const classname: string = classnames('builder-card', {
      [`${className}`]: className,
    })

    const style: CSSProperties = {
      flex: `0 0 ${widthPixels}px`,
    }

    return (
      <>
        <div
          className={classname}
          data-testid={testID}
          style={style}
          ref={this.cardRef}
        >
          {children}
        </div>
        <Handle onDrag={this.handleDrag} />
      </>
    )
  }

  private handleDrag = (mouseX: number): void => {
    if (!this.cardRef.current) {
      return
    }

    const {left, width} = this.cardRef.current.getBoundingClientRect()

    const updatedWidth = this.state.widthPixels + mouseX - (left + width)
    const enforcedWidth = Math.min(
      Math.max(updatedWidth, BUILDER_CARD_MIN_WIDTH),
      BUILDER_CARD_MAX_WIDTH
    )

    this.setState({widthPixels: enforcedWidth})
  }
}
