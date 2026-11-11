--[[
    This will fetch everything that can change when an update is pushed to Oaklands.

    All of it is fetched at once to reduce the amount of running tasks,
    especially with the rate limits.
]]

_G.Server = {
    IsOnStaging = false
}

local PublicAPIUtility = require(game.ServerScriptService.PublicAPIUtility)
PublicAPIUtility:Initialize()

local Translations = require(game.ServerScriptService.PublicAPIUtility.Translations)
local Changelogs = require(game.ServerScriptService.PublicAPIUtility.Changelogs)
local Newsletters = require(game.ServerScriptService.PublicAPIUtility.Newsletters)
local MaterialStockMarket = require(game.ServerScriptService.PublicAPIUtility.MaterialStockMarket)
local DailyMaterialBuffs = require(game.ServerScriptService.PublicAPIUtility.DailyMaterialBuffs)
local Items = require(game.ServerScriptService.PublicAPIUtility.Items)
local StoreItems = require(game.ServerScriptService.PublicAPIUtility.StoreItems)
local Minimap = require(game.ServerScriptService.PublicAPIUtility.Minimap)
local Index = require(game.ServerScriptService.PublicAPIUtility.Index)
local Regions = require(game.ServerScriptService.PublicAPIUtility.Regions)

local BinaryOutputBuffer = PublicAPIUtility:EncodeBinaryOutput({
    Translations = Translations,
    Changelogs = Changelogs,
    Newsletters = Newsletters,
    StockMarket = MaterialStockMarket,
    DailyMaterialBuffs = DailyMaterialBuffs,
    ItemDetails = Items,
    StoreItems = StoreItems,
    Minimap = Minimap,
    Index = Index,
    Regions = Regions,
})

local Result: LuauExecutionTaskOutput = {
    BinaryOutput = BinaryOutputBuffer,
}

return Result