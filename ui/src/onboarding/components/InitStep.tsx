// Libraries
import React, {PureComponent} from 'react'
// Components
import {ErrorHandling} from 'src/shared/decorators/errors'

// Types
import {StepStatus} from 'src/clockface/constants/wizard'
import {Button, ComponentColor, ComponentSize} from '@influxdata/clockface'
import {OnboardingStepProps} from 'src/onboarding/containers/OnboardingWizard'

@ErrorHandling
class InitStep extends PureComponent<OnboardingStepProps> {
  public render() {
    return (
      <>
        <h1 className="cf-funnel-page--title" data-testid="init-step--head-main">
          Welcome to InfluxDB 2.0
        </h1>
        <p
          className="cf-funnel-page--subtitle"
          data-testid="init-step--head-sub"
        >
          Get started in just a few easy steps
        </p>
        <Button
          color={ComponentColor.Primary}
          text="Get Started"
          size={ComponentSize.Large}
          onClick={this.handleNext}
          testID="onboarding-get-started"
        />
      </>
    )
  }

  private handleNext = () => {
    const {
      onSetStepStatus,
      currentStepIndex,
      onIncrementCurrentStepIndex,
    } = this.props
    onSetStepStatus(currentStepIndex, StepStatus.Complete)
    onIncrementCurrentStepIndex()
  }
}

export default InitStep
