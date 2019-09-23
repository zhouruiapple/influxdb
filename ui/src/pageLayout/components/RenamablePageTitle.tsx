// Libraries
import React, {
  PureComponent,
  KeyboardEvent,
  ChangeEvent,
  MouseEvent,
} from 'react'
import classnames from 'classnames'

// Components
import {Input, Page, Icon, IconFont} from '@influxdata/clockface'
import {ClickOutside} from 'src/shared/components/ClickOutside'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

interface Props {
  onRename: (name: string) => void
  onClickOutside?: (e: MouseEvent<HTMLElement>) => void
  name: string
  placeholder: string
  maxLength: number
  subTitle?: string
}

interface State {
  isEditing: boolean
  workingName: string
}

@ErrorHandling
class RenamablePageTitle extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props)

    this.state = {
      isEditing: false,
      workingName: props.name,
    }
  }

  public render() {
    const {name, placeholder} = this.props
    const {isEditing} = this.state

    if (isEditing) {
      return (
        <div className={this.className}>
          <ClickOutside onClickOutside={this.handleStopEditing}>
            {this.input}
          </ClickOutside>
          {this.subTitle}
        </div>
      )
    }

    return (
      <div className={this.className}>
        <div className={this.titleClassName} onClick={this.handleStartEditing}>
          <Page.Title title={name || placeholder} />
          <Icon glyph={IconFont.Pencil} />
        </div>
        {this.subTitle}
      </div>
    )
  }

  private get className(): string {
    const {subTitle} = this.props

    return classnames('renamable-page-title', {
      'renamable-page-title__subtitle': subTitle,
    })
  }

  private get subTitle(): JSX.Element {
    const {subTitle} = this.props
    if (subTitle) {
      return <Page.SubTitle title={subTitle} />
    }
  }

  private get input(): JSX.Element {
    const {placeholder, maxLength} = this.props
    const {workingName} = this.state

    return (
      <Input
        maxLength={maxLength}
        autoFocus={true}
        spellCheck={false}
        placeholder={placeholder}
        onFocus={this.handleInputFocus}
        onChange={this.handleInputChange}
        onKeyDown={this.handleKeyDown}
        className="renamable-page-title--input"
        value={workingName}
      />
    )
  }

  private handleStartEditing = (): void => {
    this.setState({isEditing: true})
  }

  private handleStopEditing = async (e): Promise<void> => {
    const {workingName} = this.state
    const {onRename, onClickOutside} = this.props

    await onRename(workingName)

    if (onClickOutside) {
      onClickOutside(e)
    }

    this.setState({isEditing: false})
  }

  private handleInputChange = (e: ChangeEvent<HTMLInputElement>): void => {
    this.setState({workingName: e.target.value})
  }

  private handleKeyDown = async (
    e: KeyboardEvent<HTMLInputElement>
  ): Promise<void> => {
    const {onRename, name} = this.props
    const {workingName} = this.state

    if (e.key === 'Enter') {
      await onRename(workingName)
      this.setState({isEditing: false})
    }

    if (e.key === 'Escape') {
      this.setState({isEditing: false, workingName: name})
    }
  }

  private handleInputFocus = (e: ChangeEvent<HTMLInputElement>): void => {
    e.currentTarget.select()
  }

  private get titleClassName(): string {
    const {name, placeholder} = this.props

    const nameIsUntitled = name === placeholder || name === ''

    return classnames('renamable-page-title--title', {
      untitled: nameIsUntitled,
    })
  }
}

export default RenamablePageTitle
