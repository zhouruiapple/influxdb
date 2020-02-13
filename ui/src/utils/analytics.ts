import amplitude from 'amplitude-js'

import {AMPLITUDE_API_KEY} from 'src/shared/constants'

class Analytics {
  private amplitude

  constructor() {
    this.amplitude = amplitude.getInstance()
    if (ENABLE_AMPLITUDE) {
      this.amplitude.init(AMPLITUDE_API_KEY)
    }
  }

  fireEvent(eventType: string): void {
    if (ENABLE_AMPLITUDE) {
      this.amplitude.logEvent(eventType)
    }
  }
}

const analytics = new Analytics()
export {analytics}
