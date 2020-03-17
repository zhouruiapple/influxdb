import React from 'react'

const INTERVAL = 120

const url = 'https://us-west-2-1.aws.cloud2.influxdata.com'
const orgId = '275ac1e8a61d71f2'
const bucketName = 'performance'
const authToken = 'iEoLyuknDOIzw7ByDlwAmTwPADgFD0C7Jni6u3UMc5iL18Ph4V6bnQex753GBv_L3yqQCMOI1yr7sJBPsYChtg=='

export class PerformanceBoy extends React.Component {
  averageFPS: number
  cummulativeFPS: number[]
  frameStartTime: number
  lastTick: number
  startTime: number
  timeInCurrentFrame: number

  constructor(props) {
    super(props)
    this.startTime = Date.now()

    this.averageFPS = 0
    this.currentFPS = 0
    this.cummulativeFPS = Array(INTERVAL)
    this.maxFPS = 0

    this.frameStartTime = Date.now()
    this.lastTick = this.frameStartTime
    this.timeInCurrentFrame = this.frameStartTime

    this.totalFrames = 0

    this.state = {
      maxTimeInFrame: 0,
      totalTimeInFrame: 0,
    }
  }

  public componentDidMount() {
    this.updateFPSCounter()
    this.writeToDatabase()
  }

  updateFPSCounter = () => {
    let {
      maxTimeInFrame,
      totalTimeInFrame,
    } = this.state
    this.totalFrames++
    this.frameStartTime = Date.now()

    this.currentFPS = this.totalFrames / (Math.floor(this.frameStartTime / 1000) - Math.floor(this.startTime / 1000))

    if (this.currentFPS > this.maxFPS) {
      this.maxFPS = this.currentFPS
    }

    this.cummulativeFPS.push(this.currentFPS)

    this.timeInCurrentFrame = Date.now() - this.frameStartTime
    totalTimeInFrame += this.timeInCurrentFrame

    if (this.timeInCurrentFrame > maxTimeInFrame) {
      maxTimeInFrame = this.timeInCurrentFrame
    }

    this.setState({
      maxTimeInFrame,
      totalTimeInFrame,
    })

    this.lastTick = Date.now()
    requestAnimationFrame(this.updateFPSCounter)
  }

  writeToDatabase = () => {
    if (this.totalFrames % INTERVAL === 0) {
      // https://v2.docs.influxdata.com/v2.0/write-data/#influxdb-api
      this.averageFPS = (this.cummulativeFPS.reduce((total, currentFPS) => (total + currentFPS), 0) / INTERVAL)

      fetch(`${url}/api/v2/write?org=${orgId}&bucket=${bucketName}&precision=s`, {
        method: 'POST',
        mode: 'cors',
        headers: {
          'Authorization': `Token ${authToken}`
        },
        body: `performance,environment=dev,interval=${INTERVAL} fps=${this.currentFPS},maxFPS=${this.maxFPS},totalFrames=${this.totalFrames},averageFrames=${this.averageFPS} ${Math.floor(Date.now() / 1000)}`
      })
      this.maxFPS = this.currentFPS
      this.cummulativeFPS = Array(INTERVAL)
    }
    requestAnimationFrame(this.writeToDatabase)
  }

  public render() {
    return (
      <>
        <div style={{ zIndex: 10000 }}>
          fps: {this.currentFPS}
          <br />
          max: {this.maxFPS}
          <br />
          totalFrames: {this.totalFrames}
          <br />
          average time per frame: {this.state.totalTimeInFrame / this.totalFrames * 1000}
          <br />
          worst time in frame: {this.state.maxTimeInFrame}
        </div>
        {this.props.children}
      </>
    )
  }
}
