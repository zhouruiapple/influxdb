// Libraries
import React, {FunctionComponent} from 'react'

// Types
import {Snippit} from 'src/types'

interface OwnProps {
  snippit: Snippit
}

type Props = OwnProps

const SnippitTooltipContents: FunctionComponent<Props> = ({
  snippit,
}) => {
  return (
    <div
      className="flux-toolbar--popover"
      data-testid="flux-toolbar--snippit-popover"
    >
      <div className="flux-function-docs--heading">
        {snippit.name}
      </div>
      <div className="flux-function-docs--snippet">
        <code>
          {snippit.code}
        </code>
      </div>
    </div>
  )
}

export default SnippitTooltipContents