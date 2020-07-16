// Libraries
import React, {FC} from 'react'

// Types
import {PipeProp} from 'src/notebooks'

// Components
import BucketSelector from 'src/notebooks/pipes/Data/BucketSelector'
import {FlexBox, ComponentSize} from '@influxdata/clockface'

// Styles
import 'src/notebooks/pipes/Query/style.scss'

const DataSource: FC<PipeProp> = ({Context}) => {
  return (
    <Context>
      <FlexBox
        margin={ComponentSize.Large}
        stretchToFitWidth={true}
        className="data-source"
      >
        <BucketSelector />
      </FlexBox>
    </Context>
  )
}

export default DataSource
