Alpine.store('sites', {
    init() {
        Alpine.store('queryresult').refresh()

        let currentObj = Alpine.store('current')
        currentObj.prepareResult = (rows, total) => {
            if (!rows) {
                return
            }
            rows.forEach(row => {
                if (row.preview) {
                    row.view_on_site = row.preview
                }
            })
        }
    }
})
