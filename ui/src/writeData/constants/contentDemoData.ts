// Constants
import {DEMO_DATA} from 'src/shared/constants/routes'

// Types
import {WriteDataItem, WriteDataSection} from 'src/writeData/constants'

// Markdown
import websiteMonitoringMarkdown from 'src/writeData/components/demoData/websiteMonitoring.md'

// Graphics
import websiteMonitoring from 'src/writeData/graphics/websiteMonitoring.svg'

export const WRITE_DATA_DEMO_DATA: WriteDataItem[] = [
  {
    id: 'website-monitoring',
    name: 'Website Monitoring',
    url: `${DEMO_DATA}/website-monitoring`,
    image: websiteMonitoring,
    markdown: websiteMonitoringMarkdown,
  },
]

export const WRITE_DATA_DEMO_DATA_SECTION: WriteDataSection = {
  id: DEMO_DATA,
  name: 'Demo Data',
  description: 'Explore our platform without having to write your own data',
  items: WRITE_DATA_DEMO_DATA,
  featureFlag: 'load-data-demo-data',
  cloudOnly: false,
}
