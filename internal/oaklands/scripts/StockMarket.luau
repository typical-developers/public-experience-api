_G.Server = {
    IsOnStaging = false
}

local PublicAPIUtility = require(game.ServerScriptService.PublicAPIUtility)
PublicAPIUtility:Initialize()

local MaterialStockMarket = require(game.ServerScriptService.PublicAPIUtility.MaterialStockMarket)

local BinaryOutputBuffer = PublicAPIUtility:EncodeBinaryOutput(MaterialStockMarket)

local Result: LuauExecutionTaskOutput = {
    BinaryOutput = BinaryOutputBuffer,
}

return Result