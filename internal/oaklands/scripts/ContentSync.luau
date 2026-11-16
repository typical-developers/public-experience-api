--[[
    This will fetch everything that can change when an update is pushed to Oaklands.

    All of it is fetched at once to reduce the amount of running tasks,
    especially with the rate limits.

    What gets fetched is decided in game by PublicAPIUtility:Export.
    A section that is missing or errors is put under `Errors` and the rest are still returned.
]]

_G.Server = {
    IsOnStaging = false
}

local PublicAPIUtility = require(game.ServerScriptService.PublicAPIUtility)
PublicAPIUtility:Initialize()

local Export = PublicAPIUtility:Export()
if typeof(Export) ~= "table" then
    error("PublicAPIUtility:Export did not return anything in this place version")
end

-- An empty table is encoded as a JSON array instead of an object, so empty fields are left out.
for _, Field in { "Sections", "Versions", "Errors" } do
    if typeof(Export[Field]) == "table" and next(Export[Field]) == nil then
        Export[Field] = nil
    end
end

local Result: LuauExecutionTaskOutput = {
    BinaryOutput = PublicAPIUtility:EncodeBinaryOutput(Export),
}

return Result
