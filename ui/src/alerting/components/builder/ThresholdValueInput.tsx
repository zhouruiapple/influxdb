// Libraries
import React, {Component, ChangeEvent} from 'react'

// Components
import {FlexBox, Input, InputType, ComponentStatus} from '@influxdata/clockface'

// Types
interface Props {
  value: number
  changeValue: (value: number) => void
}

interface State {
  workingValue: string
  inputStatus: ComponentStatus
}

class ThresholdValueStatement extends Component<Props, State> {
  constructor(props: Props) {
    super(props)

    this.state = {
      workingValue: `${this.props.value}`,
      inputStatus: ComponentStatus.Valid,
    }
  }

  componentDidUpdate() {
    if (`${this.props.value}` !== this.state.workingValue) {
      this.setState({workingValue: `${this.props.value}`})
    }
  }

  render() {
    const {workingValue, inputStatus} = this.state

    return (
      <FlexBox.Child testID="component-spacer--flex-child">
        <Input
          onChange={this.handleInputChange}
          onBlur={this.handleInputBlur}
          name=""
          testID="input-field"
          type={InputType.Text}
          value={workingValue}
          status={inputStatus}
        />
      </FlexBox.Child>
    )
  }

  private handleInputChange = (e: ChangeEvent<HTMLInputElement>): void => {
    // const {value} = this.props
    const workingValue = e.target.value.replace(/[^0-9.]/gi, '')

    let inputStatus

    if (workingValue === '') {
      inputStatus = ComponentStatus.Error
    } else {
      inputStatus = ComponentStatus.Valid
    }

    console.log(workingValue)

    this.setState({inputStatus, workingValue})
    // if (workingValue !== `${value}`) {
    // } else {
    //   this.setState({workingValue})
    // }
  }

  private handleInputBlur = (e: ChangeEvent<HTMLInputElement>): void => {
    const {changeValue} = this.props
    if (e.target.value !== '') {
      changeValue(Number(e.target.value))
    }
  }
}

export default ThresholdValueStatement
