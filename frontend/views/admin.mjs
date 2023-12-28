// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// NanoMon Frontend
// ----------------------------------------------------------------------------

import { config } from '../app.mjs'

export const adminComponent = (api) => ({
  apiEndpoint: config.API_ENDPOINT,
  message: '',
  error: '',

  async exportMonitors() {
    const monitors = await api.getMonitors()

    // Remove some fields we don't want to export
    for (const mon of monitors) {
      delete mon.id
      delete mon.updated
    }

    // Force a download of the JSON response
    const a = document.createElement('a')
    a.href = `data:application/json,${encodeURIComponent(JSON.stringify(monitors))}`
    a.download = 'nanomon-export.json'
    a.click()
  },

  async importMonitors() {
    console.log('### Importing monitors...')
    this.error = ''
    this.message = ''

    const file = document.getElementById('importBtn').files[0]
    if (!file) {
      return
    }

    const reader = new FileReader()
    reader.onload = async (e) => {
      try {
        const data = JSON.parse(e.target.result)
        await api.importMonitors(data)

        this.message = `Import successful. ${data.length} monitor(s) were imported`
      } catch (err) {
        console.error(err)
        this.error = `Import failed. ${err}`
      }
    }

    reader.readAsText(file)
  },
})