Alpine.store('page', {
    async loadCategories() {
        const path = Alpine.store('objects').find(obj => /Category/i.test(obj.name)).path
        let resp = await fetch(`${path}`, {
            method: 'POST', body: '{}'
        })
        let data = await resp.json()
        return data.items || []
    },
    init() {
        let currentObj = Alpine.store('current')
        if (!currentObj.listMode) {
            currentObj.listMode = 'grid'
        }

        currentObj.prepareEdit = (editobj, isCreate, row) => {
            if (isCreate) {
                editobj.names.id.value = randText(12)
            }

            let oldDoSave = editobj.doSave
            editobj.doSave = (event, closeWhenDone = true) => {
                oldDoSave.call(editobj, event, closeWhenDone)
                Alpine.store('editobj').names.is_draft.value = true
            }

            editobj.Published = (event, value) => {
                event.preventDefault()
                let eo = Alpine.store('editobj')
                let published = eo.names.published
                let is_draft = eo.names.is_draft

                currentObj.doAction({
                    path: `${currentObj.path}make_publish`,
                    onDone: () => {
                        is_draft.value = false
                        published.value = true
                        published.dirty = true
                        Alpine.store('queryresult').refresh()
                    },
                }, [eo.primaryValue]).then()
            }

            //TODO: REMOVE this CODE when we have a better way to handle this
            editobj.names.category_id.category_path = editobj.names.category_path

            editobj.names.draft.textareaRows = 25
            editobj.doMarkPublished = (event, value) => {
                let published = Alpine.store('editobj').names.published
                published.value = value
                published.dirty = true
            }
        }

        currentObj.prepareResult = async (rows, total) => {
            if (!rows) {
                return
            }
            let categories = await this.loadCategories()
            rows.forEach(row => {
                let categoryIdCol = row.cols.find(col => col.name == 'category_id')
                categoryIdCol.categories = categories
                categoryIdCol.category_path = row.rawData['category_path']
                row.view_on_site = currentObj.buildApiUrl(row)
                let names = {}
                row.cols.forEach(col => {
                    names[col.name] = col
                })
                row.names = names
            })
        }
        Alpine.store('queryresult').refresh()
        injectFrom('result_form_grid', 'list_page_grid.html')
    },
})