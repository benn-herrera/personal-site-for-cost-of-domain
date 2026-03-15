-- me-story.lua
-- Pandoc Lua filter for generating foldout sections in HTML from Markdown
--
-- Transforms:
--   ## intro {.intro}         → <section class="intro"> (h2 suppressed)
--   ## heading {#id}          → <section id="id"><h2>heading</h2>
--   ::: lede                  → <p class="lede">...</p>
--   ::: story / ::: {.story .open}
--                             → <div class="story"><details[open]>
--                                  <summary class="story-lead">
--                                      <div class="meta">…</div><h3>…</h3>
--                                  </summary> body </details></div>
--   ::: ai-disclosure         → <p class="ai-disclosure">...</p>

local stringify = pandoc.utils.stringify

-- Render a single block to an HTML string via pandoc's writer.
local function block_to_html(block)
    return pandoc.write(pandoc.Pandoc({ block }), "html")
end

-- Process one block within a section, returning a list of blocks.
local function process_block(block)
    if block.t ~= "Div" then
        return { block }
    end

    local classes = {}
    for _, c in ipairs(block.classes) do classes[c] = true end

    -- ::: lede
    if classes["lede"] then
        if #block.content == 1 and block.content[1].t == "Para" then
            local html = block_to_html(block.content[1])
            html = html:gsub("^<p>", '<p class="lede">', 1)
            return { pandoc.RawBlock("html", html) }
        end
        return { block }
    end

    -- ::: ai-disclosure
    if classes["ai-disclosure"] then
        if #block.content == 1 and block.content[1].t == "Para" then
            local html = block_to_html(block.content[1])
            html = html:gsub("^<p>", '<p class="ai-disclosure">', 1)
            return { pandoc.RawBlock("html", html) }
        end
        return { block }
    end

    -- ::: story / ::: {.story .open}
    if classes["story"] then
        local is_open = classes["open"]
        local meta_text, h3_text = nil, nil
        local body = {}

        for _, b in ipairs(block.content) do
            -- First pure-italic para → meta
            if not meta_text and b.t == "Para"
                and #b.content == 1 and b.content[1].t == "Emph" then
                meta_text = stringify(b.content[1])
                -- First h3 → story title
            elseif not h3_text and b.t == "Header" and b.level == 3 then
                h3_text = stringify(b)
            else
                table.insert(body, b)
            end
        end

        local open_attr = is_open and " open" or ""
        local open_html = string.format(
            '<div class="story">\n<details%s>\n<summary class="story-lead">\n' ..
            '    <div class="meta">%s</div>\n    <h3>%s</h3>\n</summary>',
            open_attr, meta_text or "", h3_text or "")

        local result = { pandoc.RawBlock("html", open_html) }
        for _, b in ipairs(body) do
            for _, pb in ipairs(process_block(b)) do
                table.insert(result, pb)
            end
        end
        table.insert(result, pandoc.RawBlock("html", "</details>\n</div>"))
        return result
    end

    return { block }
end

-- Top-level filter: wraps h2-delimited groups in <section> elements.
function Pandoc(doc)
    local new_blocks = {}
    local blocks = doc.blocks
    local i = 1

    while i <= #blocks do
        local b = blocks[i]

        if b.t == "Header" and b.level == 2 then
            local id = b.identifier
            local classes = {}
            for _, c in ipairs(b.classes) do classes[c] = true end

            -- Section open tag: intro uses class only, others use id.
            local open_tag
            if classes["intro"] then
                open_tag = '<section class="intro">'
            else
                open_tag = '<section id="' .. id .. '">'
            end
            table.insert(new_blocks, pandoc.RawBlock("html", open_tag))

            -- Emit h2 for all sections except intro (no visible heading there).
            if not classes["intro"] then
                -- New header without id — id lives on the <section> above.
                table.insert(new_blocks, pandoc.Header(2, b.content))
            end

            i = i + 1

            -- Collect blocks until the next h2.
            while i <= #blocks
                and not (blocks[i].t == "Header" and blocks[i].level == 2) do
                for _, pb in ipairs(process_block(blocks[i])) do
                    table.insert(new_blocks, pb)
                end
                i = i + 1
            end

            table.insert(new_blocks, pandoc.RawBlock("html", "</section>"))
        else
            table.insert(new_blocks, b)
            i = i + 1
        end
    end

    return pandoc.Pandoc(new_blocks, doc.meta)
end
