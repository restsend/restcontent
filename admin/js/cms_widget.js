const CategoryItemIcons = {
    Remove: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
    <path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
  </svg>
  `,
    Up: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
  <path stroke-linecap="round" stroke-linejoin="round" d="M12 19.5v-15m0 0l-6.75 6.75M12 4.5l6.75 6.75" />
</svg>
`,
    Down: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
    <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m0 0l6.75-6.75M12 19.5l-6.75-6.75" />
  </svg>
  `,
    Left: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
  <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 12h-15m0 0l6.75 6.75M4.5 12l6.75-6.75" />
</svg>
`,
    Right: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
<path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12h15m0 0l-6.75-6.75M19.5 12l-6.75 6.75" />
</svg>
`,
    Checked: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
<path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
</svg>`
}

function injectFrom(id, url) {
    let elm = document.getElementById(id)
    if (!elm) {
        return
    }
    fetch(url, {
        method: 'GET',
        cache: "no-store",
    }).then((resp) => {
        if (!resp.ok) {
            return
        }
        resp.text().then((text) => {
            elm.innerHTML = text
        })
    })
}

function formatSizeHuman(size) {
    if (!size) {
        return '0 byte'
    }
    if (size < 1024) {
        return size + ' bytes'
    }
    size = size / 1024
    if (size < 1024) {
        return size.toFixed(2) + ' KB'
    }
    size = size / 1024
    if (size < 1024) {
        return size.toFixed(2) + ' MB'
    }
    size = size / 1024
    return size.toFixed(2) + ' GB'
}

function randText(length = 8) {
    let result = 'j'
    for (let i = 0; i < length + 1; i++) {
        const padding = result.length < length ? length - result.length : 0
        result += Math.random().toString(36).substring(2, 2 + padding)
    }
    return result
}

class CategoryItem {
    constructor({ path, name, icon, children }) {
        this.path = path
        this.name = name
        this.icon = icon
        this.children = children
        this.el = undefined
    }
}

class CategoryItemWidget extends window.AdminWidgets.struct {
    render(elm) {
        let items = this.field.value || []
        let node = document.createElement('div')
        node.className = 'flex space-x-1 items-center overflow-x-hidden'
        items.forEach((item) => {
            const color = item.name.startsWith('$') ? 'cyan' : 'blue'
            let tag = document.createElement('div')
            tag.className = `flex items-center space-x-1.5 justify-between px-1.5 py-1 rounded-md text-xs bg-${color}-100 text-${color}-600`
            tag.innerText = item.name || item.path
            node.appendChild(tag)
        })
        elm.appendChild(node)
    }

    createItemElm(item) {
        let itemElm = document.createElement('div')
        itemElm.className = 'flex space-x-1'
        let inputGroup = document.createElement('div')
        inputGroup.className = 'flex flex-grow space-x-2'

        let inputName = document.createElement('input')
        inputName.className = "flex-1 rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
        inputName.type = 'text'
        inputName.value = item.name
        inputName.placeholder = 'Name'

        inputName.addEventListener('change', (e) => {
            item.name = e.target.value
            this.field.dirty = true
        })
        inputGroup.appendChild(inputName)

        let inputPath = document.createElement('input')
        inputPath.className = "flex-1 rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
        inputPath.type = 'text'
        inputPath.value = item.path
        inputName.placeholder = 'Path'

        inputPath.addEventListener('change', (e) => {
            item.path = e.target.value
            this.field.dirty = true
        })
        inputGroup.appendChild(inputPath)
        inputGroup.appendChild(inputName)

        itemElm.appendChild(inputGroup)

        {
            let btnUp = document.createElement('button')
            btnUp.className = "w-6 h-6 inline-flex items-center mt-1 px-1 py-2 border border-transparent text-xs font-medium rounded text-indigo-700 bg-indigo-100 hover:bg-indigo-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            btnUp.innerHTML = CategoryItemIcons.Up
            btnUp.addEventListener('click', (e) => {
                let index = this.field.value.findIndex(i => i.path === item.path)
                if (index > 0) {
                    itemElm.remove()
                    // remove item
                    this.field.value.splice(index, 1)
                    this.field.value.splice(index - 1, 0, item)
                    this.field.dirty = true
                    this.itemsGroup.insertBefore(itemElm, this.itemsGroup.children[index - 1])
                }
            })
            itemElm.appendChild(btnUp)
        }
        {
            let btnDown = document.createElement('button')
            btnDown.className = "w-6 h-6 inline-flex items-center mt-1 px-1 py-2 border border-transparent text-xs font-medium rounded text-indigo-700 bg-indigo-100 hover:bg-indigo-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            btnDown.innerHTML = CategoryItemIcons.Down
            btnDown.addEventListener('click', (e) => {
                let index = this.field.value.findIndex(i => i.path === item.path)
                if (index >= 0 && index < this.field.value.length - 1) {
                    // remove item
                    this.field.value.splice(index, 1)
                    this.field.value.splice(index + 1, 0, item)
                    this.field.dirty = true
                    itemElm.remove()
                    this.itemsGroup.insertBefore(itemElm, this.itemsGroup.children[index + 1])
                }
            })
            itemElm.appendChild(btnDown)
        }
        {
            let btnRemove = document.createElement('button')
            btnRemove.className = "w-6 h-6 inline-flex items-center mt-1 px-1 py-2 border border-transparent text-xs font-medium rounded text-indigo-700 bg-indigo-100 hover:bg-indigo-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            btnRemove.innerHTML = CategoryItemIcons.Remove
            btnRemove.addEventListener('click', (e) => {
                itemElm.remove()
                this.field.value = this.field.value.filter(i => i.path !== item.path)
                this.field.dirty = true
            })
            itemElm.appendChild(btnRemove)
        }

        this.itemsGroup.appendChild(itemElm)
    }

    renderEdit(elm) {
        if (!this.field.value) {
            this.field.value = []
        }
        this.$el = elm
        let node = document.createElement('div')
        node.className = 'flex flex-col w-full h-96 bg-gray-50 rounded shadow overflow-scroll px-4 py-4'
        {
            let addGroup = document.createElement('div')
            addGroup.className = 'flex space-x-2'
            let addInput = document.createElement('input')
            addInput.className = "flex-grow rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
            addInput.type = 'text'
            addInput.placeholder = 'Add category'
            let addButton = document.createElement('span')
            addButton.className = "inline-flex items-center px-2.5 py-2 border border-transparent text-xs font-medium rounded text-indigo-700 bg-indigo-100 hover:bg-indigo-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 cursor-pointer"
            addButton.innerText = 'Add'

            addGroup.appendChild(addInput)
            addGroup.appendChild(addButton)

            addButton.addEventListener('click', (e) => {
                let name = addInput.value
                addInput.value = ''
                if (name) {
                    let item = new CategoryItem({ path: randText(), name })
                    this.field.value.push(item)
                    this.field.dirty = true
                    this.createItemElm(item)
                }
            })
            node.appendChild(addGroup)
        }

        {

            let labels = document.createElement('div')
            labels.className = 'flex pt-2 text-xs font-medium text-gray-500'
            labels.innerHTML = `<div class="w-60">Path</div><div class="flex-1">Name</div>`
            node.appendChild(labels)

            let itemsGroup = document.createElement('div')
            itemsGroup.className = 'mt-2 flex flex-col space-y-2'
            this.itemsGroup = itemsGroup
            node.appendChild(itemsGroup)
        }
        for (let item of this.field.value) {
            this.createItemElm(item)
        }
        elm.appendChild(node)
    }
}

class HumanizeSizeWidget extends window.AdminWidgets.string {
    render(elm) {
        const size = this.field.value || 0
        this.renderWith(elm, formatSizeHuman(size))
    }
}
class IsDraftWidget extends window.AdminWidgets.bool {
    render(elm) {
        if (this.field.value === true) {
            const cls = "text-center rounded-md whitespace-nowrap px-2 py-1 text-xs font-medium ring-1 ring-inset text-yellow-800 bg-yellow-50 ring-yellow-600/20"
            elm.innerHTML = `<p class="${cls}"> Draft</p>`
        }
    }
}
class TagsWidget extends window.AdminWidgets.string {
    static splitTags(text) {
        return text.split(/[;,，；]+/g).filter((v) => v).map((v) => v.trim())
    }

    static async loadExistsTags(path) {
        let resp = await fetch(`${path}tags`, { method: 'POST' })
        let items = await resp.json() || []
        let tags = []

        items.forEach(item => {
            //ignore case filter exists items
            let items = TagsWidget.splitTags(item)
            items = items.filter((v) => tags.findIndex((exist) => exist.toLowerCase() == v.toLowerCase()) == -1)
            tags.push(...items)
        })
        return tags
    }

    createShowTag(text, close = false) {
        const color = text.startsWith('$') ? 'cyan' : 'blue'
        let tag = document.createElement('div')
        tag.className = `flex items-center space-x-1.5 justify-between px-1.5 py-1 rounded-md text-xs bg-${color}-100 text-${color}-600`

        let textTag = document.createElement('span')
        textTag.className = ``
        textTag.innerText = text
        tag.appendChild(textTag)

        if (close) {
            let closeTag = document.createElement('button')
            closeTag.className = `inline-flex items-center justify-center w-4 h-4 rounded-full bg-${color}-200 hover:bg-${color}-300 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-${color}-500`
            closeTag.innerHTML = `<span class="sr-only">Remove</span>
            <svg class="h-2 w-2 text-${color}-600" stroke="currentColor" fill="none" viewBox="0 0 8 8">
                <path stroke-linecap="round" d="M1 1l6 6m0-6L1 7" />
            </svg>`
            closeTag.addEventListener('click', (e) => {
                e.preventDefault()
                this.tags.splice(this.tags.findIndex((v) => v.toLowerCase() === text.toLowerCase()), 1)

                this.field.value = this.tags.join(',')
                this.field.dirty = true

                this.options.querySelectorAll(`li`).forEach(item => {
                    if (item.innerText.trim() === text) {
                        item.querySelector('div span').classList.toggle('hidden')
                    }
                })
                tag.remove()
            })
            tag.appendChild(closeTag)
        }
        return tag
    }


    render(elm) {
        let tags = TagsWidget.splitTags(this.field.value || '')
        let node = document.createElement('div')
        node.className = 'flex space-x-1 items-center overflow-x-hidden flex-wrap gap-x-2 gap-y-1'

        tags.forEach((v) => {
            node.appendChild(this.createShowTag(v))
        })
        elm.appendChild(node)
    }

    renderEdit(elm) {
        this.tags = TagsWidget.splitTags(this.field.value || '')
        let node = document.createElement('div')
        node.className = 'flex px-2 py-2 items-center w-full flex-wrap gap-x-2 gap-y-1 rounded border border-gray-300 focus-within:ring-1 focus-within:ring-indigo-500 focus-within:border-indigo-500'
        this.tagsNode = document.createElement('div')
        this.tagsNode.className = 'flex flex-wrap items-center gap-x-2 gap-y-1'

        this.tags.forEach((v) => {
            this.tagsNode.appendChild(this.createShowTag(v, true))
        })

        node.appendChild(this.tagsNode)
        let appendDiv = document.createElement('div')
        appendDiv.className = 'inline-flex group relative'
        let appendButton = document.createElement('div')
        appendDiv.appendChild(appendButton)
        appendButton.className = 'hover:cursor-pointer z-20'
        appendButton.innerHTML = `<div class="flex rounded-full items-center justify-center w-6 h-6 bg-indigo-200 hover:bg-indigo-300 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
        <svg class="h-4 w-4 text-indigo-600" stroke="currentColor" fill="none" viewBox="0 0 8 8">
            <path stroke-linecap="round" d="M4 1v6m3-3H1" />
        </svg></div>`
        // drop down menu
        let menu = document.createElement('div') // 
        menu.className = 'hidden group-hover:block left-0 absolute pt-8 ease-in-out delay-150 bg-opacity-0 z-10 transform px-2 w-screen max-w-md'
        menu.innerHTML = `<div class="bg-white py-6 rounded-lg shadow-lg ring-1 ring-black ring-opacity-5 overflow-hidden">
        <div class="flex px-5 justify-between gap-x-4 items-center">
            <input type="text" class="flex-1 rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" placeholder="Add tag">
            <button class="justify-center px-4 py-1.5 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none">
                Add
            </button>
        </div>        
        <div>
            <ul class="mx-4 py-4 grid grid-cols-3">        
            </ul>
        </div>
        </div>`

        this.options = menu.getElementsByTagName('ul')[0]
        let button = menu.getElementsByTagName('button')[0]
        button.addEventListener('click', (e) => {
            e.preventDefault()
            let input = menu.getElementsByTagName('input')[0]
            let text = input.value
            input.value = ''

            if (text) {
                if (this.tags.findIndex((v) => v.toLowerCase() === text.toLowerCase()) != -1) {
                    return
                }
                let existElm = undefined
                this.options.querySelectorAll(`li`).forEach(item => {
                    if (item.innerText.trim().toLowerCase() === text) {
                        existElm = item
                        existElm.querySelector('div span').classList.toggle('hidden')
                    }
                })

                this.tags.push(text)
                this.field.value = this.tags.join(',')
                this.field.dirty = true

                if (!existElm) {
                    this.options.appendChild(this.createTagSelectItem(text))
                }
                this.tagsNode.appendChild(this.createShowTag(text, true))
            }
        })

        this.tags.forEach((v) => {
            this.options.appendChild(this.createTagSelectItem(v))
        })

        TagsWidget.loadExistsTags(Alpine.store('current').path).then((items) => {
            items.forEach((text) => {
                if (this.tags.findIndex((v) => v.toLowerCase() === text.toLowerCase()) != -1) {
                    return
                }
                this.options.appendChild(this.createTagSelectItem(text))
            })
        })

        appendDiv.appendChild(menu)
        node.appendChild(appendDiv)
        elm.appendChild(node)
    }

    createTagSelectItem(text) {
        const checked = this.tags.findIndex((v) => v.toLowerCase() === text.toLowerCase()) >= 0
        let item = document.createElement('li')
        item.className = `py-2 hover:bg-gray-100 hover:cursor-pointer`
        item.innerHTML = `<div class="flex items-center">
            <div class="mx-4 text-sm font-medium text-blue-900 w-8 h-6">
            <span class="${checked ? '' : 'hidden'}">${checked ? CategoryItemIcons.Checked : ''}</span>
            </div>
            <div class="flex-1 flex-shrink-0">${text}</div>
        </div>`
        item.addEventListener('click', (e) => {
            e.preventDefault()
            const isChecked = this.tags.findIndex((v) => v.toLowerCase() === text.toLowerCase()) >= 0
            if (isChecked) {
                this.tags.splice(this.tags.findIndex((v) => v.toLowerCase() === text.toLowerCase()), 1)
                // query span contains text
                this.tagsNode.querySelectorAll(`div > span`).forEach(elm => {
                    if (elm.innerText.trim().toLowerCase() === text.toLowerCase()) {
                        elm.parentElement.remove()
                    }
                })
            } else {
                this.tags.push(text)
                this.tagsNode.appendChild(this.createShowTag(text, true))
            }
            item.querySelector('div span').classList.toggle('hidden')

            this.field.value = this.tags.join(',')
            this.field.dirty = true
        })
        return item
    }
}

class CategoryWidget extends window.AdminWidgets.string {
    static async loadCategories() {
        const path = Alpine.store('objects').find(obj => /Category/i.test(obj.name)).path
        let resp = await fetch(`${path}`, {
            method: 'POST', body: '{}'
        })
        let data = await resp.json()
        return data.items || []
    }

    render(elm) {
        if (!this.field.value) {
            return
        }

        const categories = this.col.categories || []
        const category = categories.find((v) => v.uuid === this.field.value)
        let text = category ? category.name : this.field.value
        if (category && this.col.category_path) {
            const p = category.items.find((v) => v.path === this.col.category_path)
            if (p) {
                text += ` > ${p.name}`
            }
        }
        super.renderWith(elm, text)
    }

    createCategoryPath(categoryPathNode, category) {
        categoryPathNode.innerHTML = ''
        let firstOption = document.createElement('option')
        firstOption.value = {}
        firstOption.innerText = 'Empty value'
        categoryPathNode.appendChild(firstOption)
        if (category && category.items) {
            category.items.forEach((item) => {
                let option = document.createElement('option')
                option.value = item.path
                option.innerText = item.name
                if (item.path == this.field.category_path.value) {
                    option.selected = true
                }
                categoryPathNode.appendChild(option)
            })
        }
    }
    renderEdit(elm) {
        let node = document.createElement('div')
        node.className = 'flex gap-x-4 items-center w-full'

        let categoriesNode = document.createElement('select')
        categoriesNode.className = 'block rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset  focus:ring-indigo-600 sm:text-sm sm:leading-6'

        let categoryPathNode = document.createElement('select')
        categoryPathNode.className = 'block rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset  focus:ring-indigo-600 sm:text-sm sm:leading-6'

        let firstOption = document.createElement('option')
        firstOption.value = {}
        firstOption.innerText = this.field.placeholder || 'Empty value'
        categoriesNode.appendChild(firstOption)

        CategoryWidget.loadCategories().then((categories) => {
            for (let category of categories) {
                let option = document.createElement('option')
                option.data = category
                option.innerText = category.name || category.uuid

                if (category.uuid == this.field.value) {
                    option.selected = true
                    this.createCategoryPath(categoryPathNode, category)
                }
                categoriesNode.appendChild(option)
            }
        })

        categoriesNode.addEventListener('change', (e) => {
            e.preventDefault()

            let category = e.target.selectedOptions[0].data
            this.field.value = category.uuid
            this.field.dirty = true

            this.field.category_path.value = undefined
            this.field.category_path.dirty = true
            this.createCategoryPath(categoryPathNode, category)
        })

        categoryPathNode.addEventListener('change', (e) => {
            e.preventDefault()
            this.field.category_path.value = e.target.value
            this.field.category_path.dirty = true
        })

        node.appendChild(categoriesNode)
        node.appendChild(categoryPathNode)
        elm.appendChild(node)
    }
}

class IsPublishedWidget extends window.AdminWidgets.bool {
    render(elm) {
        let color = "green"
        let text = "Published"
        if (this.field.value === false) {
            color = "yellow"
            text = "Not published"
        }

        const cls = `inline-flex items-center gap-x-1.5 rounded-md bg-${color}-100 px-2 py-1 text-xs font-medium text-${color}-700 ring-1 ring-inset ring-${color}-200 / 20`
        elm.innerHTML = `<span class="${cls}">
            <svg class="h-1.5 w-1.5 fill-${color}-500" viewBox="0 0 6 6" aria-hidden="true">
                <circle cx="3" cy="3" r="3" />
            </svg>${text}</span> `
    }
}

window.AdminWidgets['humanize-size'] = HumanizeSizeWidget
window.AdminWidgets['category-item'] = CategoryItemWidget
window.AdminWidgets['is-draft'] = IsDraftWidget
window.AdminWidgets['is-published'] = IsPublishedWidget
window.AdminWidgets['tags'] = TagsWidget
window.AdminWidgets['category-id-and-path'] = CategoryWidget


class TagsFilterWidget extends window.AdminFilterWidgets.select {
    render(elm) {
        let options = [{ label: 'Empty value', value: '', op: '=' }]

        const oldOnSelect = this.field.onSelect
        this.field.onSelect = (field, selected) => {
            if (selected && selected.value) {
                selected.op = 'like'
                selected.showOp = 'contains'
            }
            oldOnSelect.call(this.field, field, selected)
        }
        TagsWidget.loadExistsTags(Alpine.store('current').path).then((items) => {
            items.forEach((text) => {
                options.push({ label: text, value: text })
            })
            super.renderWithOptions(elm, options, true)
        })
    }
}

class CategoryFilterWidget extends window.AdminFilterWidgets.select {
    render(elm) {
        let options = [{ label: 'Empty value', value: {}, op: '=' }]

        const oldOnSelect = this.field.onSelect
        this.field.onSelect = (field, selected) => {
            if (selected) {
                let filters = []
                if (selected.value instanceof Array) {
                    filters.push({ name: 'category_id', op: selected.op, value: selected.value.map((v) => v.id).filter(v => v) })
                    filters.push({ name: 'category_path', op: selected.op, value: selected.value.map((v) => v.path).filter(v => v) })
                } else {
                    if (selected.value.id !== undefined) {
                        filters.push({ name: 'category_id', op: selected.op, value: selected.value.id })
                    }
                    if (selected.value.path !== undefined) {
                        filters.push({ name: 'category_path', op: selected.op, value: selected.value.path })
                    }
                }
                selected.isGroup = true
                selected.value = filters
            }
            oldOnSelect.call(this.field, field, selected)
        }

        CategoryWidget.loadCategories().then((categories) => {
            categories.forEach((category) => {
                options.push({ label: category.name, value: { id: category.uuid, path: undefined } })
                let items = category.items || []
                items.forEach((item) => {
                    options.push({ label: `${category.name} > ${item.name}`, value: { id: category.uuid, path: item.path } })
                })
            })
            super.renderWithOptions(elm, options, true)
        })

    }
}

window.AdminFilterWidgets['tags'] = TagsFilterWidget
window.AdminFilterWidgets['category-id-and-path'] = CategoryFilterWidget


Alpine.directive('admin-copyclip', (el, { expression }, { evaluate }) => {
    let value = evaluate(expression)
    let node = document.createElement('button')
    node.className = "group relative ml-2 hidden h-9 w-9 items-center justify-center sm:flex"
    node.innerHTML = `<div
        class="hidden -mt-16 absolute bg-gray-800 text-white font-semibold rounded-md shadow text-sm px-2 py-1" >
            Copied
    </div>
            <svg class="h-8 w-8 stroke-slate-400 transition group-hover:rotate-[-4deg] group-hover:stroke-slate-600"
                fill="none" viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg" stroke-width="1.5"
                stroke-linecap="round" stroke-linejoin="round">
                <path
                    d="M12.9975 10.7499L11.7475 10.7499C10.6429 10.7499 9.74747 11.6453 9.74747 12.7499L9.74747 21.2499C9.74747 22.3544 10.6429 23.2499 11.7475 23.2499L20.2475 23.2499C21.352 23.2499 22.2475 22.3544 22.2475 21.2499L22.2475 12.7499C22.2475 11.6453 21.352 10.7499 20.2475 10.7499L18.9975 10.7499"
                    stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
                <path
                    d="M17.9975 12.2499L13.9975 12.2499C13.4452 12.2499 12.9975 11.8022 12.9975 11.2499L12.9975 9.74988C12.9975 9.19759 13.4452 8.74988 13.9975 8.74988L17.9975 8.74988C18.5498 8.74988 18.9975 9.19759 18.9975 9.74988L18.9975 11.2499C18.9975 11.8022 18.5498 12.2499 17.9975 12.2499Z"
                    stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
                <path d="M13.7475 16.2499L18.2475 16.2499" stroke-width="1.5" stroke-linecap="round"
                    stroke-linejoin="round"></path>
                <path d="M13.7475 19.2499L18.2475 19.2499" stroke-width="1.5" stroke-linecap="round"
                    stroke-linejoin="round"></path>
                <g class="opacity-0">
                    <path d="M15.9975 5.99988L15.9975 3.99988" stroke-width="1.5" stroke-linecap="round"
                        stroke-linejoin="round"></path>
                    <path d="M19.9975 5.99988L20.9975 4.99988" stroke-width="1.5" stroke-linecap="round"
                        stroke-linejoin="round"></path>
                    <path d="M11.9975 5.99988L10.9975 4.99988" stroke-width="1.5" stroke-linecap="round"
                        stroke-linejoin="round"></path>
                </g>
            </svg>`
    node.addEventListener('click', (e) => {
        navigator.clipboard.writeText(value).then(() => {
            // toggle hidden 
            let tooltip = node.getElementsByTagName('div')[0]
            tooltip.classList.remove('hidden')
            setTimeout(() => {
                tooltip.classList.add('hidden')
            }, 1500)
        })
    })
    el.appendChild(node)
})

Alpine.directive('admin-json-editor', (el, { expression }, { evaluate }) => {
    let field = evaluate(expression)
    let node = document.createElement('div')
    node.className = "w-full h-[40rem]"
    el.appendChild(node)

    let editor = new JSONEditor(node, {
        mode: 'code',
        onChange: () => {
            field.value = editor.getText()
            field.dirty = true
        }
    })
    editor.setText(field.value)
})

Alpine.directive('admin-markdown-editor', (el, { expression }, { evaluate }) => {
    let field = evaluate(expression)
    let node = document.createElement('textarea')
    node.rows = field.textareaRows || 3
    node.className = 'block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6'
    node.value = field.value
    node.placeholder = field.placeholder || ''
    el.appendChild(node)

    let easyMDE = new EasyMDE({ node })
    easyMDE.codemirror.on("change", () => {
        field.value = easyMDE.value()
        field.dirty = true
    })
})

Alpine.directive('admin-html-editor', (el, { expression }, { evaluate }) => {
    let field = evaluate(expression)
    let node = document.createElement('textarea')

    node.rows = field.textareaRows || 3
    node.className = 'block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6'
    node.value = field.value
    node.placeholder = field.placeholder || ''

    el.appendChild(node)

    node.addEventListener('change', (e) => {
        field.value = node.value
        field.dirty = true
    })
    const editor = Jodit.make(node, {
        height: '500',

        uploader: {
            url: './media/upload?created=true', // This is a required parameter
            isSuccess: function (resp) {
                return resp.error == undefined
            },
            getMessage: function (resp) {
                return resp.error
            },
            buildData: function (formData) {
                let file = formData.getAll('files[0]')[0]
                formData.delete('files[0]')
                formData.append('file', file)
                return formData
            },

            process: function (resp) {
                let files = [resp.publicUrl]
                let data = {
                    files,
                    path: resp.path,
                    error: resp.error,
                }
                return data
            },
            defaultHandlerSuccess: function (data) {
                data.files.forEach(f => {
                    this.s.insertImage(f)
                })
            },
        },

        filebrowser: {
            isSuccess: function (resp) {
                return resp.error == undefined
            },
            getMessage: function (resp) {
                return resp.error
            },
            ajax: {
                url: './media/',
                contentType: 'application/json',
                processData: true,
                prepareData: function (data) {
                    let path = data.path ? data.path : '/'
                    if (!path.startsWith('/')) {
                        path = `/${path}`
                    }
                    path = path.replace(/\/+/g, '/')
                    let body = {
                        filters: [
                            { name: 'path', op: '=', value: path },
                        ]
                    }
                    body.path = path
                    return body
                },
                process: function (resp) {
                    const items = resp.items || []
                    const files = items.filter((item) => !item.directory).map((item) => {
                        let type = item.content_type
                        let thumb = item.thumbnail
                        let file = item.public_url
                        let thumbIsAbsolute = true
                        let fileIsAbsolute = true

                        return {
                            type,
                            file,
                            fileIsAbsolute,
                            thumb,
                            thumbIsAbsolute,
                            name: item.name,
                            type: item.content_type,
                            size: formatSizeHuman(item.size),
                        }
                    })

                    let folders = items.filter((item) => item.directory).map((item) => item.name)
                    const path = editor.filebrowser.state.currentPath || '/'
                    if (path != '/') {
                        folders.unshift('..')
                    }
                    return {
                        success: true,
                        data: {
                            sources: [
                                {
                                    path,
                                    files,
                                    folders
                                }
                            ],
                            code: 220
                        }
                    }
                }
            },
        },
    })
    editor.value = field.value
})