_G.Server = {
    IsOnStaging = false
}

local PublicAPIUtility = require(game.ServerScriptService.PublicAPIUtility)
PublicAPIUtility:Initialize()

local DailyMaterialBuffs = require(game.ServerScriptService.PublicAPIUtility.DailyMaterialBuffs)

local BinaryOutputBuffer = PublicAPIUtility:EncodeBinaryOutput(DailyMaterialBuffs)

local Result: LuauExecutionTaskOutput = {
    BinaryOutput = BinaryOutputBuffer,
}

return Result