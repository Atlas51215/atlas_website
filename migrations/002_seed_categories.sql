-- Reviews group (4 categories with extra_fields)
INSERT INTO categories (slug, name, group_name, extra_fields, sort_order) VALUES
    ('movies', 'Movie Reviews', 'reviews',
     '[{"name":"rating","type":"float","label":"Rating","min":0,"max":10},{"name":"director","type":"text","label":"Director"},{"name":"release_year","type":"int","label":"Release Year"}]',
     1),
    ('games', 'Game Reviews', 'reviews',
     '[{"name":"rating","type":"float","label":"Rating","min":0,"max":10},{"name":"platform","type":"text","label":"Platform"},{"name":"developer","type":"text","label":"Developer"}]',
     2),
    ('tv', 'TV Reviews', 'reviews',
     '[{"name":"rating","type":"float","label":"Rating","min":0,"max":10},{"name":"seasons","type":"int","label":"Seasons"},{"name":"network","type":"text","label":"Network"}]',
     3),
    ('products', 'Product Reviews', 'reviews',
     '[{"name":"rating","type":"float","label":"Rating","min":0,"max":10},{"name":"price_range","type":"text","label":"Price Range"},{"name":"purchase_link","type":"text","label":"Purchase Link"}]',
     4);

-- Blog group (7 categories, no extra_fields)
INSERT INTO categories (slug, name, group_name, extra_fields, sort_order) VALUES
    ('movies',  'Movies',   'blog', NULL, 1),
    ('games',   'Games',    'blog', NULL, 2),
    ('tv',      'TV',       'blog', NULL, 3),
    ('products','Products', 'blog', NULL, 4),
    ('general', 'General',  'blog', NULL, 5),
    ('dev',     'Dev',      'blog', NULL, 6),
    ('tech',    'Tech',     'blog', NULL, 7);
