import { z } from '@hono/zod-openapi';

const ShopItem = z.object({
    name: z.string(),
    currency: z.string(),
    price: z.number(),
    type: z.string(),
    image: z.nullable(z.string()).optional(),
});

export type ShopItemchema = z.infer<typeof ShopItem>;

export default ShopItem;