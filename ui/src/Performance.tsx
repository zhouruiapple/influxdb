import React from 'react'

export class PerformanceBoy extends React.Component {
  frameStartTime: number
  currentFrameTime: number
  public componentDidMount() {
    this.updateFPSCounter()
    this.writeToDatabase()
  }

  updateFPSCounter = () => {
    let {
      currentFps,
      lastTick,
      maxFps,
      maxTimeInFrame,
      minFps,
      totalFrames,
      totalTimeInFrame,
    } = this.state
    this.frameStartTime = Date.now()
    currentFps = 1 / ((this.frameStartTime - lastTick) / 1000)

    if (currentFps > maxFps) {
      maxFps = currentFps
    }
    if (currentFps < minFps) {
      minFps = currentFps
    }

    totalFrames++
    lastTick = Date.now()
    this.currentFrameTime = lastTick - this.frameStartTime
    totalTimeInFrame += this.currentFrameTime
    if (this.currentFrameTime > maxTimeInFrame) {
      maxTimeInFrame = this.currentFrameTime
    }
    this.setState({
      currentFps,
      lastTick,
      maxFps,
      maxTimeInFrame,
      minFps,
      totalFrames,
      totalTimeInFrame,
    })
    requestAnimationFrame(this.updateFPSCounter)
  }

  writeToDatabase = () => {
    if (this.state.totalFrames % 300 === 0) {
      // https://v2.docs.influxdata.com/v2.0/write-data/#influxdb-api
      console.log('curl -XPOST "http://localhost:9999/api/v2/write?org=YOUR_ORG&bucket=YOUR_BUCKET&precision=s" \
        --header "Authorization: Token YOURAUTHTOKEN" \
        --data-raw "mem,host=host1 used_percent=23.43234543 1556896326"')
    }
    requestAnimationFrame(this.writeToDatabase)
  }

  constructor(props) {
    super(props)

    this.frameStartTime = Date.now()
    this.currentFrameTime = Date.now()
    this.state = {
      currentFps: 0,
      lastTick: Date.now(),
      maxFps: 0,
      maxTimeInFrame: 0,
      minFps: 100,
      totalFrames: 0,
      totalTimeInFrame: 0,
    }
  }

  public render() {
    return (
      <>
        <div style={{ zIndex: 10000 }}>
          fps: {this.state.currentFps}
          <br />
          max: {this.state.maxFps}
          <br />
          min: {this.state.minFps}
          <br />
          totalFrames: {this.state.totalFrames}
          <br />
          average time per frame: {this.state.totalTimeInFrame / this.state.totalFrames * 1000}
          <br />
          worst time in frame: {this.state.maxTimeInFrame}
        </div>
        {this.props.children}
      </>
    )
  }
}
