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

local Changelogs = require(game.ServerScriptService.PublicAPIUtility.Changelogs)
local Newsletters = require(game.ServerScriptService.PublicAPIUtility.Newsletters)
local MaterialStockMarket = require(game.ServerScriptService.PublicAPIUtility.MaterialStockMarket)

local BinaryOutputBuffer = PublicAPIUtility:EncodeBinaryOutput({
    Changelogs = Changelogs,
    Newsletters = Newsletters,
    StockMarket = MaterialStockMarket
})

local Result: LuauExecutionTaskOutput = {
    BinaryOutput = BinaryOutputBuffer,
}

return Result