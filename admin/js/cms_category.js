
Alpine.store('category', {
    init() {
        let currentObj = Alpine.store('current')
        currentObj.prepareEdit = (editobj, isCreate, row) => {
            if (isCreate) {
                editobj.names.uuid.value = randText(10)
            }
        }

        currentObj.prepareResult = (rows, total) => {
            if (!rows) {
                return
            }
            rows.forEach(row => {
                row.view_on_site = currentObj.buildApiUrl(row)
            })
        }
        Alpine.store('queryresult').refresh()
    },
})
