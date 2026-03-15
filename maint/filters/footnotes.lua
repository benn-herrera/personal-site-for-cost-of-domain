-- footnotes.lua
-- Adds anchor links to [.LABEL] custom footnotes in dev articles.
--
-- In the source markdown:
--   Inline reference:  text[.foo]      (bracket-dot-label, no space before bracket)
--   Definition:        [.foo] Note...  (paragraph starting with bracket-dot-label)
--
-- LABEL can be any non-empty string without "]": symbols (†), numbers (1), words (foo), etc.
-- IDs are derived directly from the label: [.foo] → id="fn-foo" / id="fnref-foo"
--
-- Phase 1 — Para: converts definition paragraphs to <p id="fn-LABEL"> with backlinks.
--            Header: strips id from ## Footnotes to avoid conflict with pandoc's section.
-- Phase 2 — Str: wraps inline [.LABEL] occurrences with <a href="#fn-LABEL"> links.
--           Handles both standalone Str "[.foo]" and embedded Str "gap[.foo]".

local function ids_for(label)
    return "fn-" .. label, "fnref-" .. label
end

local function anchor_link(label)
    local fn_id, ref_id = ids_for(label)
    return pandoc.RawInline("html", string.format(
        '<a href="#%s" id="%s" role="doc-noteref"><sup>%s</sup></a>',
        fn_id, ref_id, label))
end

-- Phase 1: definition paragraphs and heading id

local phase1 = {
    Header = function(el)
        -- Strip id from ## Footnotes — pandoc also generates <section id="footnotes">
        if el.level == 2 and el.identifier == "footnotes" then
            el.identifier = ""
            return el
        end
    end,

    Para = function(el)
        -- Definition paragraphs start with a [.LABEL] Str token.
        if #el.content == 0 then return nil end
        local first = el.content[1]
        if first.t ~= "Str" then return nil end
        local label = first.text:match("^%[%.([^%]]+)%]$")
        if not label then return nil end
        local fn_id, ref_id = ids_for(label)

        -- Build paragraph content without the [.LABEL] prefix and following space
        local content = {}
        local start = 2
        if #el.content >= 2 and el.content[2].t == "Space" then
            start = 3
        end
        for i = start, #el.content do
            table.insert(content, el.content[i])
        end

        local html = pandoc.write(pandoc.Pandoc({pandoc.Para(content)}), "html")
        local backlink = string.format(
            '<a href="#%s" class="footnote-back" role="doc-backlink">%s&#8202;↩︎</a>&#160;',
            ref_id, label)
        html = html:gsub("^<p>", '<p id="' .. fn_id .. '">' .. backlink, 1)
        return pandoc.RawBlock("html", html)
    end,
}

-- Phase 2: inline [.LABEL] references (handles "[.foo]" embedded in a Str)

local phase2 = {
    Str = function(el)
        local s, e, label = el.text:find("%[%.([^%]]+)%]")
        if s then
            local result = {}
            if s > 1 then
                table.insert(result, pandoc.Str(el.text:sub(1, s - 1)))
            end
            table.insert(result, anchor_link(label))
            local tail = el.text:sub(e + 1)
            if tail ~= "" then
                table.insert(result, pandoc.Str(tail))
            end
            return result
        end
    end,
}

return {phase1, phase2}
