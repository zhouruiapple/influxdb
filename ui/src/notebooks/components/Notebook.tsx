import React, {FC} from 'react'

import {Page} from '@influxdata/clockface'
import 'src/notebooks/Notebook.scss'

const NotebookPage: FC = () => {
  return (
    <Page titleTag="Notebook">
      <Page.Header fullWidth={false} />
      <Page.Contents fullWidth={false} scrollable={true} />
    </Page>
  )
}

export default NotebookPage
