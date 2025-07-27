-- This will fetch everything that can change with an update being pushed to Oaklands.
-- 
-- All of it is fetched at once to reduce the amount of task requests made,
-- especially considering the potential rate limits that could be exhausted.
-- 
-- This script is ONLY ran when the update time for Oaklands changes.
local HttpService = game:GetService("HttpService")

_G.Server = {
    IsOnStaging = false
}

local Translations = require(game.ReplicatedFirst.Client.GUI.Translations).new() -- Available to entire script since some script need translation values.
Translations:ClientLoaded()

type BinaryInputData = {
    CachedNewsletters: {{ string: boolean }},
    CachedChangelogs: {{ string: boolean }},
}

local TaskInput: LuauExecutionTaskInput = ({...})[1]
local BinaryInput: buffer = TaskInput.BinaryInput
local Input: BinaryInputData = HttpService:JSONDecode(
    buffer.readstring(BinaryInput, 0, buffer.len(BinaryInput))
)

local Worlds = require(game.ServerScriptService.Worlds)
for _,v in pairs(game.Workspace.Worlds:GetChildren()) do
    if (v.Name == "Shared" or v.Name == Worlds.CurrentWorld) then
        for _,Inst in pairs(v:GetChildren()) do
            if (Inst:IsA("PackageLink")) then
                continue
            end
            Inst.Parent = game.Workspace.World
        end
    end
end
game.Workspace.Worlds:Destroy()

function GetMaterialStockMarket()
    local MaterialValues = require(game.ServerScriptService.Server.Economy.MaterialValues).new()

    local TreesV2 = require(game.ServerScriptService.Server.Generation.TreesV2).new()
    local Rocks = require(game.ServerScriptService.Server.Generation.Rocks).new()

    Rocks:PreloadVariants()
    Rocks:PreloadOres()
    TreesV2:PreloadSpecies()

    function MaterialValues:UpdateAllDisplays()
        return
    end

    function MaterialValues:GetClassFromType(Material: string, Type: string)
        local Class = nil

        if (Type == "Trees") then
            Class = TreesV2
            Class = Class and Class.TreeSpecies
            Class = Class and Class:GetSpeciesModule(Material)
        end
        
        if (Type == "Rocks") then
            Class = Rocks
            Class = Class and Class.RockVariants
            Class = Class and Class:GetVariantModule(Material)
        end

        if (Type == "Ores") then
            Class = Rocks
            Class = Class and Class.RockOres
            Class = Class and Class:GetOreModule(Material)
        end

        if (not Class) then
            return nil
        end
        
        return Class
    end

    MaterialValues:UpdateMaterialValues()

    MaterialValues.LastRandom = Random.new(MaterialValues.MaterialValuesLastSeed)
    MaterialValues.CurrentRandom = Random.new(MaterialValues.MaterialValuesUpdatedSeed)

    return MaterialValues:CreateFormattedValues()
end

function GetNewsletterPages()
    local Newsletter = require(game.ReplicatedFirst.Newsletter)
    local Pages = game.ReplicatedFirst.Newsletter:GetChildren()
    
    local Details = {
        latest_page = Newsletter.LatestNewsletter,
        pages = {}
    }

    local FormatContent = function (PageContent)
        local Contents = {}

        for i, content in pairs (PageContent) do
            if (content.Type == "Image") then
                Contents[i] = {
                    type = content.Type,
                    image_id = content.ImageId
                }
            end

            if (content.Type == "Video") then
                Contents[i] = {
                    type = content.Type,
                    video_id = content.VideoId
                }
            end

            if (content.Type == "Paragraph") then
                Contents[i] = {
                    type = content.Type,
                    text = content.Text
                }
            end

            if (content.Type == "ImageCarousel") then
                Contents[i] = {
                    type = content.Type,
                    image_ids = content.ImageIds
                }
            end
        end

        return Contents
    end

    local FormatSections = function (PageSections)
        local Sections = {}

        for i, section in pairs(PageSections) do
            local Contents = FormatContent(section.Content)

            Sections[i] = {
                header = section.Header,
                contents = Contents
            }
        end

        return Sections
    end

    for _, page in pairs(Pages) do
        local Title = page.Name
        if (Title == "_Types" or Input.CachedNewsletters[Title]) then
            continue
        end

        local Info = require(page)
        Details.pages[Title] = {
            banner_image_id = Info.BannerImageId,
            header_text = Info.Header,
            subheader_text = Info.SubHeader,
            sections = FormatSections(Info.Sections)
        }
    end

    if not next(Details.pages) then
        Details.pages = nil
    end

    return Details
end

function GetChangelogs()
    local Changelog = require(game.ReplicatedFirst.Changelog)

    local Versions = {}

    for i, version in pairs(Changelog.versions) do
        if (Input.CachedChangelogs[i]) then
            continue
        end

        Versions[i] = {
            _id = version.Id,
            date = version.Date,
            changed = version.Changed or {},
            added = version.Added or {},
            fixed = version.Fixed or {},
        }
    end

    if not next(Versions) then
        return nil
    end

    return Versions
end

function GetItemDetails()
    local Content = game.ServerScriptService.Server.Content
    local Objects = Content.Object:GetChildren()

    local ItemDetails = {}

    for _, object in pairs(Objects) do
        if (object.Name == "Behavior" or string.sub(object.Name, 1, 1) == "_") then
            continue
        end

        local Items = Content.Object:FindFirstChild(object.Name):GetChildren()
        for _, item in pairs(Items) do
            if (string.sub(item.Name, 1, 1) == "_") then
                continue
            end

            local TranslatedName = Translations:GetTranslation(item.Name)
            local TranslatedDescription = Translations:GetTranslation(`{item.Name}_description`) or ""

            ItemDetails[item.Name] = {}
            local ItemInfo = ItemDetails[item.Name]

            ItemInfo.details = {
                identifier = item.Name,
                name = TranslatedName,
                description = TranslatedDescription,
            }

            -- Fetches any store information about the item.
            if (item:FindFirstChild("Store")) then
                local Form = require(item.Store).new()
                local TranslatedStoreName = Translations:GetTranslation(Form.StoreName)
                local Currency = Form.CustomCurrency or "Cash"
                ItemInfo.store = {
                    currency = Currency,
                    price = Form.Price,
                    type = Form.ConvertType:lower(),
                }
            end

            -- Fetches any stat item information about the item.
            if (item:FindFirstChild("Stats")) then
                if (item:FindFirstChild("Stats")) then
                    local Stats = require(item.Stats)
                    ItemInfo.item = Stats
                end
            end

            -- Fetches any gift information about the item.
            if (item:FindFirstChild("Gift")) then
                local Form = require(item.Gift).new()
                ItemInfo.gift = {
                    unbox_epoch = Form.UnboxTime,
                }
            end
        end
    end

    return ItemDetails
end

function GetTranslations()
    return Translations.LoadedLanguages
end

function GetStoreItems()
    local Stores = game.Workspace.World.Regions.Stores:GetChildren()
    local StoreItems = {}

    -- Removes duplicate children.
    local RemoveDupes = function(items)
        local Items = {}
        local ItemCheck = {}

        for _, item in (items) do
            local ItemName = item.Name

            if (item:GetAttribute("StagingOnly")) then
                continue
            end
            if (ItemCheck[ItemName]) then
                continue
            end

            ItemCheck[ItemName] = true
            table.insert(Items, ItemName)
        end

        return Items
    end

    -- Removes any special characters and spaces.
    local FormatStoreName = function(name)
        return name:gsub("[^%w]", "")
    end

    for _, store in pairs(Stores) do
        local StoreName = FormatStoreName(store.Name)

        -- Fetches the classic store and sets the data manually.
        -- A seed needs to be calculated first to get the current items.
        if (store.Name == "Classic Store") then
            StoreItems[StoreName] = {}

            local ClassicStore = require(game.ServerScriptService.Server.World.Stores.Finlay["Classic Store"]).new()
            ClassicStore:CalculateSeed()

            local AvailableItems = ClassicStore.CurrentItems
            if (#AvailableItems <= 0) then
                continue
            end

            for _, item in AvailableItems do
                table.insert(StoreItems["ClassicStore"], item)
            end

            continue
        end
        -- 

        local Items = store:FindFirstChild("Items")
        if (not Items) then
            continue
        end

        local AvailableItems = RemoveDupes(Items:GetChildren())
        if (#AvailableItems <= 0) then
            continue
        end

        if (not StoreItems[StoreName]) then
            StoreItems[StoreName] = {}
        end

        for _, item in AvailableItems do
            table.insert(StoreItems[StoreName], item)
        end
    end

    return StoreItems
end

function GetOreRarity_V1()
    local RockRNG = require(game.ServerScriptService.Server.Generation.Rocks.Variants._RockClass.RNG)
    local RNG = {}

    local FormatKey = function(input: string)
        return input
            :gsub("(%u)", "_%1")
            :gsub("%s+", "_")
            :gsub("^_", "")
            :gsub("__+", "_")
            :lower()
    end

    for k, rocks in pairs(RockRNG) do
        if (k == "corrupted_limestone") then
            continue
        end

        local RockKey = FormatKey(k)

        if (not RNG[RockKey]) then
            RNG[RockKey] = {}
        end

        for k, region in pairs(rocks) do
            local RegionKey = FormatKey(k)

            if (not RNG[RockKey][RegionKey]) then
                RNG[RockKey][RegionKey] = {}
            end

            for k, data in pairs(region) do
                if (k == false) then
                    RNG[RockKey][RegionKey][k] = nil
                    RNG[RockKey][RegionKey]["_none"] = data
                    continue
                end

                local dataKey = FormatKey(k)
                RNG[RockKey][RegionKey][dataKey] = data
            end
        end
    end

    return RNG
end

local UpdatedData = {
    translations = GetTranslations(),
    newsletters = GetNewsletterPages(),
    changelogs = GetChangelogs(),
    item_details = GetItemDetails(),
    stock_market = GetMaterialStockMarket(),
    store_items = GetStoreItems(),

    ore_rarity_v1 = GetOreRarity_V1(),
}

local EncodedUpdateData = HttpService:JSONEncode(UpdatedData)
local Buff = buffer.create(#EncodedUpdateData)
buffer.writestring(Buff, 0, EncodedUpdateData)

local Result: LuauExecutionTaskOutput = {
    BinaryOutput = Buff,
}

return Result