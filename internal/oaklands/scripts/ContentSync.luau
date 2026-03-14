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
local Items = require(game.ServerScriptService.PublicAPIUtility.Items)
local StoreItems = require(game.ServerScriptService.PublicAPIUtility.StoreItems)

local BinaryOutputBuffer = PublicAPIUtility:EncodeBinaryOutput({
    Translations = Translations,
    Changelogs = Changelogs,
    Newsletters = Newsletters,
    StockMarket = MaterialStockMarket,
    ItemDetails = Items,
    StoreItems = StoreItems
})

local Result: LuauExecutionTaskOutput = {
    BinaryOutput = BinaryOutputBuffer,
}

return Result