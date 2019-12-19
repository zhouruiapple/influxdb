// Libraries
import React, {PureComponent, ChangeEvent} from 'react'
import {getDeep} from 'src/utils/wrappers'

// Components
import {Form, Input, Grid} from '@influxdata/clockface'
import OnboardingButtons from 'src/onboarding/components/OnboardingButtons'

// Actions
import {setupAdmin} from 'src/onboarding/actions'

// Types
import {ISetupParams} from '@influxdata/influx'
import {
  Columns,
  IconFont,
  InputType,
  ButtonType,
  ComponentSize,
  ComponentStatus,
  ComponentColor,
  Bullet,
  Panel,
  CTAButton,
} from '@influxdata/clockface'
import {StepStatus} from 'src/clockface/constants/wizard'
import {OnboardingStepProps} from 'src/onboarding/containers/OnboardingWizard'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

interface State extends ISetupParams {
  confirmPassword: string
  isPassMismatched: boolean
}

interface Props extends OnboardingStepProps {
  onSetupAdmin: typeof setupAdmin
}

@ErrorHandling
class AdminStep extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props)
    const {setupParams} = props

    const username = getDeep(setupParams, 'username', '')
    const password = getDeep(setupParams, 'password', '')
    const confirmPassword = getDeep(setupParams, 'password', '')
    const org = getDeep(setupParams, 'org', '')
    const bucket = getDeep(setupParams, 'bucket', '')

    this.state = {
      username,
      password,
      confirmPassword,
      org,
      bucket,
      isPassMismatched: false,
    }
  }

  public render() {
    const {
      username,
      password,
      confirmPassword,
      org,
      bucket,
      isPassMismatched,
    } = this.state
    const icon = this.InputIcon
    const status = this.InputStatus
    return (
      <>
        <Form onSubmit={this.handleNext}>
          <h1
            className="cf-funnel-page--title"
            data-testid="admin-step--head-main"
          >
            Welcome to InfluxDB!
          </h1>
          <p
            className="cf-funnel-page--subtitle"
            data-testid="admin-step--head-sub"
          >
            Before using the platform, configure the first <strong>User, Organization, and Bucket.</strong><br/>You will be able to create additional Users, Organizations and
            Buckets later.
          </p>
          <Grid>
            <Grid.Row>
              <Grid.Column
                widthXS={Columns.Twelve}
                widthMD={Columns.Eight}
                offsetMD={Columns.Two}
              >
                <Panel>
                  <Panel.SymbolHeader symbol={<Bullet text="1" />} title={<h5 className="cf-funnel-page--panel-title">Initial User</h5>} />
                  <Panel.Body>
                    <Grid>
                      <Grid.Row>
                        <Grid.Column widthXS={Columns.Twelve}>
                          <Form.Element label="Username">
                            <Input
                              value={username}
                              onChange={this.handleUsername}
                              titleText="Username"
                              size={ComponentSize.Medium}
                              icon={icon}
                              status={status}
                              disabledTitleText="Username has been set"
                              autoFocus={true}
                              testID="input-field--username"
                            />
                          </Form.Element>
                        </Grid.Column>
                        <Grid.Column widthXS={Columns.Six}>
                          <Form.Element label="Password">
                            <Input
                              type={InputType.Password}
                              value={password}
                              onChange={this.handlePassword}
                              titleText="Password"
                              size={ComponentSize.Medium}
                              icon={icon}
                              status={status}
                              disabledTitleText="Password has been set"
                              testID="input-field--password"
                            />
                          </Form.Element>
                        </Grid.Column>
                        <Grid.Column widthXS={Columns.Six}>
                          <Form.Element
                            label="Confirm Password"
                            errorMessage={
                              isPassMismatched && 'Passwords do not match'
                            }
                          >
                            <Input
                              type={InputType.Password}
                              value={confirmPassword}
                              onChange={this.handleConfirmPassword}
                              titleText="Confirm Password"
                              size={ComponentSize.Medium}
                              icon={icon}
                              status={this.passwordStatus}
                              disabledTitleText="password has been set"
                              testID="input-field--password-chk"
                            />
                          </Form.Element>
                        </Grid.Column>
                      </Grid.Row>
                    </Grid>
                  </Panel.Body>
                </Panel>
                <Panel>
                  <Panel.SymbolHeader symbol={<Bullet text="2" />} title={<h5 className="cf-funnel-page--panel-title">Initial Organization & Bucket</h5>} />
                  <Panel.Body>
                    <Grid>
                      <Grid.Row>
                        <Grid.Column widthXS={Columns.Six}>
                          <Form.Element
                            label="Organization Name"
                            testID="form-elem--orgname"
                          >
                            <Input
                              value={org}
                              onChange={this.handleOrg}
                              titleText="Organization Name"
                              size={ComponentSize.Medium}
                              icon={icon}
                              status={ComponentStatus.Default}
                              placeholder="ex: DevOps team"
                              disabledTitleText="Initial organization name has been set"
                              testID="input-field--orgname"
                            />
                          </Form.Element>
                          <p className="onboarding--explainer">An organization is a workspace for a group of users requiring access to time series data, dashboards, and other resources.
        You can create organizations for different functional groups, teams, or projects.</p>
                        </Grid.Column>
                        <Grid.Column widthXS={Columns.Six}>
                          <Form.Element
                            label="Bucket Name"
                            testID="form-elem--bucketname"
                          >
                            <Input
                              value={bucket}
                              onChange={this.handleBucket}
                              titleText="Bucket Name"
                              size={ComponentSize.Medium}
                              icon={icon}
                              status={status}
                              placeholder="ex: My Cool Bucket"
                              disabledTitleText="Initial bucket name has been set"
                              testID="input-field--bucketname"
                            />
                          </Form.Element>
                          <p className="onboarding--explainer">A bucket is where your time series data is stored with a retention policy.</p>
                        </Grid.Column>
                      </Grid.Row>
                    </Grid>
                  </Panel.Body>
                </Panel>
                <p className="cf-funnel-page--subtitle">Next: <strong>Quick Start Options</strong></p>
                <CTAButton text="Continue" color={ComponentColor.Primary} disabledTitleText="Complete the form above to continue" status={this.nextButtonStatus} type={ButtonType.Submit} />
              </Grid.Column>
            </Grid.Row>
          </Grid>
          {/* <OnboardingButtons
            nextButtonStatus={this.nextButtonStatus}
            autoFocusNext={false}
          /> */}
        </Form>
      </>
    )
  }

  private get isAdminSet(): boolean {
    const {stepStatuses, currentStepIndex} = this.props
    if (stepStatuses[currentStepIndex] === StepStatus.Complete) {
      return true
    }
    return false
  }

  private handleUsername = (e: ChangeEvent<HTMLInputElement>): void => {
    const username = e.target.value
    this.setState({username})
  }

  private handlePassword = (e: ChangeEvent<HTMLInputElement>): void => {
    const {confirmPassword} = this.state
    const password = e.target.value
    const isPassMismatched = confirmPassword && password !== confirmPassword
    this.setState({password, isPassMismatched})
  }

  private handleConfirmPassword = (e: ChangeEvent<HTMLInputElement>): void => {
    const {password} = this.state
    const confirmPassword = e.target.value
    const isPassMismatched = confirmPassword && password !== confirmPassword
    this.setState({confirmPassword, isPassMismatched})
  }

  private handleOrg = (e: ChangeEvent<HTMLInputElement>): void => {
    const org = e.target.value
    this.setState({org})
  }

  private handleBucket = (e: ChangeEvent<HTMLInputElement>): void => {
    const bucket = e.target.value
    this.setState({bucket})
  }

  private get nextButtonStatus(): ComponentStatus {
    if (this.areInputsValid) {
      return ComponentStatus.Default
    }
    return ComponentStatus.Disabled
  }

  private get areInputsValid(): boolean {
    const {
      username,
      password,
      org,
      bucket,
      confirmPassword,
      isPassMismatched,
    } = this.state

    return (
      username &&
      password &&
      confirmPassword &&
      org &&
      bucket &&
      !isPassMismatched
    )
  }

  private get passwordStatus(): ComponentStatus {
    const {isPassMismatched} = this.state
    if (this.isAdminSet) {
      return ComponentStatus.Disabled
    }
    if (isPassMismatched) {
      return ComponentStatus.Error
    }
    return ComponentStatus.Default
  }

  private get InputStatus(): ComponentStatus {
    if (this.isAdminSet) {
      return ComponentStatus.Disabled
    }
    return ComponentStatus.Default
  }

  private get InputIcon(): IconFont {
    if (this.isAdminSet) {
      return IconFont.Checkmark
    }
    return null
  }

  private handleNext = async () => {
    const {
      onIncrementCurrentStepIndex,
      onSetupAdmin: onSetupAdmin,
      onSetStepStatus,
      currentStepIndex,
    } = this.props

    const {username, password, org, bucket} = this.state

    if (this.isAdminSet) {
      onSetStepStatus(currentStepIndex, StepStatus.Complete)
      onIncrementCurrentStepIndex()
      return
    }

    const setupParams = {
      username,
      password,
      org,
      bucket,
    }

    const isAdminSet = await onSetupAdmin(setupParams)
    if (isAdminSet) {
      onIncrementCurrentStepIndex()
    }
  }
}

export default AdminStep
