// Libraries
import React, {PureComponent, MouseEvent} from 'react'
import classnames from 'classnames'

interface Props {
  onDrag: (mouseX: number) => void
}

interface State {
  isDragging: boolean
}

export default class ResizableBuilderCardHandle extends PureComponent<
  Props,
  State
> {
  public state: State = {
    isDragging: false,
  }

  public render() {
    const {isDragging} = this.state

    const classname: string = classnames('builder-card--handle', {
      'builder-card--handle__dragging': isDragging,
    })

    return <div className={classname} onMouseDown={this.handleStartDrag} />
  }

  private handleStartDrag = (): void => {
    this.setState({isDragging: true})

    window.addEventListener('mouseup', this.handleStopDrag)
    window.addEventListener('mousemove', this.handleDrag)
  }

  private handleStopDrag = (): void => {
    this.setState({isDragging: false})

    window.removeEventListener('mouseup', this.handleStopDrag)
    window.removeEventListener('mousemove', this.handleDrag)
  }

  private handleDrag = (e: MouseEvent<HTMLDivElement>): void => {
    const {onDrag} = this.props

    onDrag(e.pageX)
  }
}
