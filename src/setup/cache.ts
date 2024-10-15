import NodeSchedule from "node-schedule";
import { getMaterialStockMarket, getCurrentClassicShop, getMaterialLeaderboard } from "@/lib/util";
import container from "@/setup/container";

const cacheRunners = {
    materialStockMarket: async () => {
        const values = await getMaterialStockMarket();
        container.cache.set('material_stock_market', values);
    },
    classicShop: async () => {
        const values = await getCurrentClassicShop();
        container.cache.set('classic_shop', values);
    },
    materialLeaderboard: async () => {
        const values = await getMaterialLeaderboard();

        const reset = new Date();
        reset.setUTCDate(reset.getUTCDate() + 1);
        reset.setUTCHours(0, 0, 0, 0);

        container.cache.set('material_leaderboard', {
            reset_time: reset,
            last_update: new Date().toISOString(),
            leaderboards: values
        });
    }
};

await Promise.all(Object.entries(cacheRunners).map(([_, func]) => func()));

// Runs every 16th hour.
NodeSchedule.scheduleJob('reset_classic_shop', '0 16 * * *', async () => 
    await Promise.all([
        cacheRunners.classicShop(),
    ])
);

// Runs every 4th, 10th, 16th, 20th hour.
NodeSchedule.scheduleJob('reset_stockmarket', '0 4,10,16,20 * * *', async () => 
    await Promise.all([
        cacheRunners.materialStockMarket()
    ])
);

// Runs every 5th minute
NodeSchedule.scheduleJob('refetch_leaderboard', '*/5 * * * *', async() => 
    await Promise.all([
        cacheRunners.materialLeaderboard()
    ])
);