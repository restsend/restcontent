Alpine.store('media', {
    listMode: 'list',
    _folders: [],
    current: '/',
    current_dirs: [],
    fileChoice: '',
    uploading: false,
    _previewUrl: '',
    _localPreview: false,
    _canPreview: false,

    get folders() {
        return this._folders
    },
    get foldersCount() {
        if (!this._folders) {
            return 0
        }
        if (this.current != '/') {
            return this._folders.length - 1
        }
        return this._folders.length
    },

    async refreshFolders() {
        this.uploading = false
        this.current = '/'
        this._localPreview = false
        this._previewUrl = ''
        this._canPreview = false
        this.current_dirs = []
        await this.changeFolder(this.current)
    },

    async createFolder(elm) {

        if (!elm.value) {
            return
        }

        let name = elm.value
        elm.value = ''

        if (this._folders.find(f => f.name === name)) {
            return
        }
        let url = `./media/new_folder?path=${this.current}&name=${name}`
        let req = await fetch(url, {
            method: 'POST',
        })
        let path = await req.json()
        this._folders.push({ path, name, foldersCount: 0, filesCount: 0 })
    },

    async changeFolder(path, event) {
        if (event) {
            event.preventDefault()
        }

        this.current = path
        let dirs = []
        let step = ''
        path.split('/').forEach((dir, i) => {
            if (i === 0) {
                dirs.push({ path: '/', name: '/' })
            } else {
                step += '/' + dir
                dirs.push({ path: step, name: dir })
            }
        })

        this.current_dirs = dirs
        let url = `./media/folders?path=${this.current}`

        let req = await fetch(url, {
            method: 'POST',
        })

        let data = await req.json()
        this._folders = []
        if (path != '/') {
            let pos = path.lastIndexOf('/')
            let parent = path.substring(0, pos)
            this._folders.push({ path: parent || '/', name: '..' })
        }

        this.current = path
        if (data) {
            this._folders.push(...data)
        }
        let query = Alpine.store('queryresult')
        query.setFilters([
            { name: 'path', value: path, op: '=' },
            { name: 'directory', value: false, op: '=' },
        ]).refresh()
    },

    choiceFile(data, file) {
        data.choicename = `${file.name} (${formatSizeHuman(file.size)})`
        data.choice = file

        if (file) {
            if (file.type.match('image.*')) {
                const reader = new FileReader();
                reader.addEventListener('load', () => {
                    this._canPreview = true
                    this._localPreview = true
                    this._previewUrl = reader.result
                    if (!Alpine.store('editobj').names.name.value) {
                        document.querySelector('#id_name > input').value = file.name
                    }
                });
                reader.readAsDataURL(file);
            } else {
                this._canPreview = false
            }
            data.showName = true
        }
    },

    async uploadFile(data, editobj) {
        if (!data.choice) {
            return
        }
        this.uploading = true
        // upload file 
        let form = new FormData()
        form.append('file', data.choice)

        let path = this.current
        let name = editobj.names.name.value || data.choice.name
        let url = `./media/upload?path=${path}&name=${name}`

        data.choice = undefined
        data.choicename = undefined
        Alpine.store('toasts').doing('Uploading file ...')

        let isOk = false
        try {
            let upload = await fetch(url, {
                method: 'POST',
                body: form
            })

            if (upload.status != 200) {
                Alpine.store('toasts').error(upload.statusText)
                this.uploading = false
                return false
            }

            let data = await upload.json()
            if (!data) {
                console.error('upload failed', upload.status, upload.statusText)
                this.uploading = false
                return false
            }

            let { storePath, dimensions, ext, size, contentType, publicUrl, external } = data
            editobj.names.store_path.value = storePath
            editobj.names.store_path.dirty = true

            editobj.names.external.value = external
            editobj.names.external.dirty = true

            editobj.names.public_url.value = publicUrl
            editobj.names.public_url.dirty = true

            editobj.names.dimensions.value = dimensions
            editobj.names.dimensions.dirty = true

            editobj.names.ext.value = ext
            editobj.names.ext.dirty = true

            editobj.names.size.value = size
            editobj.names.size.dirty = true

            editobj.names.name.value = name
            editobj.names.name.dirty = true

            editobj.names.updated_at.value = new Date().toISOString()
            editobj.names.updated_at.dirty = true

            editobj.names.content_type.value = contentType
            editobj.names.content_type.dirty = true
            Alpine.store('toasts').reset()
            isOk = true
        } catch (e) {
            Alpine.store('toasts').error(e.toString())
            isOk = false
        }
        this.uploading = false
        return isOk
    },

    async doSave(event, data) {
        let editobj = Alpine.store('editobj')
        await this.uploadFile(data, editobj)
        await editobj.doSave(event, false)
    },

    async doCreate(event, data) {
        let editobj = Alpine.store('editobj')
        const isOk = await this.uploadFile(data, editobj)
        if (isOk) {
            await editobj.doSave(event, false)
        }
    },

    onRemove(dir, event) {
        Alpine.store('confirmAction').confirm({
            action: {
                name: 'Remove',
                title: 'Remove directory',
                class: 'text-white bg-red-500 hover:bg-red-700',
                path: 'media/remove_dir',
                text: `<p>Remove directory <strong>${dir}<strong></p> ?
                <strong class="text-red-500">This will remove all files in this directory!</strong>`,
                onDone: (keys, result) => {
                    if (!result) {
                        result = '/'
                    }
                    this.changeFolder(result).then()
                }
            }, keys: [
                { path: dir },
            ]
        })
    },
    get formatSize() {
        const editobj = Alpine.store('editobj')
        let size = editobj.names.size && editobj.names.size.value || 0
        return formatSizeHuman(size)
    },
    get formatDate() {
        const editobj = Alpine.store('editobj')
        if (!editobj.names.updated_at || editobj.names.updated_at.value == '') {
            return '-'
        }
        let date = editobj.names.updated_at.value
        let d = new Date(date)
        return d.toLocaleDateString()
    },
    get formatPixel() {
        const editobj = Alpine.store('editobj')
        return editobj.names.dimensions && editobj.names.dimensions.value || '-'
    },
    get formatExt() {
        const editobj = Alpine.store('editobj')
        return editobj.names.ext && editobj.names.ext.value || '-'
    },

    get formatUrl() {
        const editobj = Alpine.store('editobj')
        return editobj.names.public_url && editobj.names.public_url.value || ''
    },

    get canPreview() {
        if (this._localPreview) {
            return this._canPreview
        }
        const editobj = Alpine.store('editobj')
        if (!editobj || !editobj.names || !editobj.names.content_type) {
            return false
        }
        if (editobj.names.content_type.value && /image/i.test(editobj.names.content_type.value)) {
            return true
        }
        return false
    },

    get isLocalPreview() {
        return this._localPreview
    },

    get previewUrl() {
        if (this._previewUrl != '') {
            return this._previewUrl
        }
        return this.formatUrl
    },

    init() {
        let currentObj = Alpine.store('current')
        if (!currentObj.listMode) {
            currentObj.listMode = 'list'
        }
        currentObj.prepareEdit = (editobj, isCreate, row) => {
            Alpine.store('media')._previewUrl = ''
            Alpine.store('media')._localPreview = false
            if (isCreate) {
                editobj.names.path.value = this.current
                editobj.names.published.value = true
            }

            editobj.doMarkPublished = (event, value) => {
                let published = Alpine.store('editobj').names.published
                published.value = value
                published.dirty = true
            }
        }

        currentObj.prepareQuery = (query, source) => {
            query.filters.push({ name: 'path', value: this.current, op: '=' })
            query.filters.push({ name: 'directory', value: false, op: '=' })
            return query
        }

        currentObj.prepareResult = (rows, total) => {
            if (!rows) {
                return
            }
            rows.forEach(row => {
                row.view_on_site = row.public_url
            })
        }
        injectFrom('result_head_form', 'media_path.html')
        injectFrom('result_form_grid', 'list_media_grid.html')
        this.refreshFolders().then()
    },
})
