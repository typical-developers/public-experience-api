_G.Server = {
    IsOnStaging = false
}

local PublicAPIUtility = require(game.ServerScriptService.PublicAPIUtility)
PublicAPIUtility:Initialize()

local ClassicShopItems = require(game.ServerScriptService.PublicAPIUtility.StoreItems.ClassicShop)

local Result: LuauExecutionTaskOutput = {
    ReturnValues = ClassicShopItems,
}

return Result