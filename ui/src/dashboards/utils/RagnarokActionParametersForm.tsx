// Libraries
import React, {PureComponent, MouseEvent} from 'react'
import {connect} from 'react-redux'

import {
    ComponentColor,
    Input,
    Button,
    Overlay,
    Form,  
  } from '@influxdata/clockface'

interface Props {
    isVisible: boolean,
    service: any,
    action: any,
    onCancel: () => void
    onApply: (obj: any) => void
}

export class RagnarokActionParametersForm extends PureComponent<Props> {

    state = {
        destinationBucket: this.props.action.output ? this.props.action.output.defaultDatabase : '',
        destinationMeasurement: this.props.action.output ? this.props.action.output.defaultMeasurement : '',
        outputTags: this.props.action.output ? this.props.action.output.defaultTags : '',
        repeat: '0s',
        parameterValues: this.getInitialParameterValues(),
    }

    public render() {
        return (
    
        <Overlay visible={this.props.isVisible}>
        <Overlay.Container maxWidth={600}>
          <Overlay.Header title={this.props.action.name + " with " + this.props.service.name}/>
          <Overlay.Body key="foo">
          {this.props.action.parameters.map(p => (
           [<Form.Label key={p.name+"-l"} required={true} label={p.name}/>,
            <Input key={p.name+"-f"} value={this.state.parameterValues[p.name]} placeholder={p.name} onChange={(e)=>{this.updateParam(p.name,e)}} />,
            <Form.HelpText key={p.name+"-h"} text={p.description}></Form.HelpText>]
          ))}
        
          <Form.Divider lineColor={ComponentColor.Primary}/>
          <Form.Label required={true} label="Destination Bucket"/>
          <Input value={this.state.destinationBucket} placeholder="Destination Bucket" onChange={this.updateDestinationBucket} />
          <Form.HelpText text="The bucket to write resuls to"/>

          <Form.Divider lineColor={ComponentColor.Primary}/>
          <Form.Label required={true} label="Destination Measurement"/>
          <Input value={this.state.destinationMeasurement} placeholder="Destination Measurement" onChange={this.updateDestinationMeasurement} />
          <Form.HelpText text="The measurement to write resuls to"/>

          <Form.Divider lineColor={ComponentColor.Primary}/>
          <Form.Label required={false} label="Output Tags"/>
          <Input value={this.state.outputTags} placeholder="Output Tags" onChange={this.updateOutputTags}/>
          <Form.HelpText text="Any additional tags to attach to the results"/>):
          
          <Form.Divider lineColor={ComponentColor.Primary}/>
          <Form.Label required={false} label="Repeat Every"/>
          <Input value={this.state.repeat} placeholder="How often to repeat" onChange={this.updateRepeat}/>
          <Form.HelpText text="If required, how often should this action to be repeated?"/>

          </Overlay.Body>
          <Overlay.Footer>
            <Button text="Apply" onClick={()=>this.props.onApply(this.state)} />
            <Button text="Cancel" onClick={this.props.onCancel} />
          </Overlay.Footer>
        </Overlay.Container>
      </Overlay>
        )
    }


private getInitialParameterValues () {
    let result = {}
    this.props.action.parameters.forEach(p =>{
        result[p.name] = p.default ? p.default : ''
    })
    return result
}

private updateParam = (name, event) => {
    let p = {}
    Object.assign(p, this.state.parameterValues );
    p[name] = event.target.value
    this.setState({parameterValues:p}) 
  }

  private updateRepeat = (event) => {
    this.setState({repeat:event.target.value})
  }

  private updateOutputTags = (event) => {
    this.setState({outputTags:event.target.value})
  }

  private updateDestinationBucket = (event) => {
    this.setState({ destinationBucket: event.target.value })
  }

  private updateDestinationMeasurement = (event) => {
    this.setState({ destinationMeasurement: event.target.value })
  }
}

  