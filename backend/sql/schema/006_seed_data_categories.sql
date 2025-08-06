-- +goose Up
/* -- unique primary categories -- */
INSERT INTO
    primary_categories (name)
VALUES
    ('INCOME'),
    ('TRANSFER_IN'),
    ('TRANSFER_OUT'),
    ('LOAN_PAYMENTS'),
    ('BANK_FEES'),
    ('ENTERTAINMENT'),
    ('FOOD_AND_DRINK'),
    ('GENERAL_MERCHANDISE'),
    ('HOME_IMPROVEMENT'),
    ('MEDICAL'),
    ('PERSONAL_CARE'),
    ('GENERAL_SERVICES'),
    ('GOVERNMENT_AND_NON_PROFIT'),
    ('TRANSPORTATION'),
    ('TRAVEL'),
    ('RENT_AND_UTILITIES')
ON CONFLICT (name) DO NOTHING;

/* -- unique detailed categories -- */
WITH
    src (name, description, primary_name) AS (
        VALUES
            /* ===========================  INCOME  =========================== */
            (
                'DIVIDENDS',
                'Dividends from investment accounts',
                'INCOME'
            ),
            (
                'INTEREST_EARNED',
                'Income from interest on savings accounts',
                'INCOME'
            ),
            (
                'RETIREMENT_PENSION',
                'Income from pension payments',
                'INCOME'
            ),
            ('TAX_REFUND', 'Income from tax refunds', 'INCOME'),
            (
                'UNEMPLOYMENT',
                'Income from unemployment benefits, including unemployment insurance and healthcare',
                'INCOME'
            ),
            (
                'WAGES',
                'Income from salaries, gig-economy work, and tips earned',
                'INCOME'
            ),
            (
                'OTHER_INCOME',
                'Other miscellaneous income, including alimony, social security, child support, and rental',
                'INCOME'
            ),
            /* ========================  TRANSFER_IN  ========================= */
            (
                'CASH_ADVANCES_AND_LOANS',
                'Loans and cash advances deposited into a bank account',
                'TRANSFER_IN'
            ),
            (
                'DEPOSIT',
                'Cash, checks, and ATM deposits into a bank account',
                'TRANSFER_IN'
            ),
            (
                'INVESTMENT_AND_RETIREMENT_FUNDS',
                'Inbound transfers to an investment or retirement account',
                'TRANSFER_IN'
            ),
            (
                'SAVINGS',
                'Inbound transfers to a savings account',
                'TRANSFER_IN'
            ),
            (
                'ACCOUNT_TRANSFER',
                'General inbound transfers from another account',
                'TRANSFER_IN'
            ),
            (
                'OTHER_TRANSFER_IN',
                'Other miscellaneous inbound transactions',
                'TRANSFER_IN'
            ),
            /* ========================  TRANSFER_OUT  ======================== */
            (
                'INVESTMENT_AND_RETIREMENT_FUNDS',
                'Transfers to an investment or retirement account, including investment apps such as Acorns, Betterment',
                'TRANSFER_OUT'
            ),
            (
                'SAVINGS',
                'Outbound transfers to savings accounts',
                'TRANSFER_OUT'
            ),
            (
                'WITHDRAWAL',
                'Withdrawals from a bank account',
                'TRANSFER_OUT'
            ),
            (
                'ACCOUNT_TRANSFER',
                'General outbound transfers to another account',
                'TRANSFER_OUT'
            ),
            (
                'OTHER_TRANSFER_OUT',
                'Other miscellaneous outbound transactions',
                'TRANSFER_OUT'
            ),
            /* ======================  LOAN_PAYMENTS  ========================= */
            (
                'CAR_PAYMENT',
                'Car loans and leases',
                'LOAN_PAYMENTS'
            ),
            (
                'CREDIT_CARD_PAYMENT',
                'Payments to a credit card. Positive for credit-card accounts; negative for checking',
                'LOAN_PAYMENTS'
            ),
            (
                'PERSONAL_LOAN_PAYMENT',
                'Personal loans, including cash advances and buy-now-pay-later repayments',
                'LOAN_PAYMENTS'
            ),
            (
                'MORTGAGE_PAYMENT',
                'Payments on mortgages',
                'LOAN_PAYMENTS'
            ),
            (
                'STUDENT_LOAN_PAYMENT',
                'Payments on student loans. For tuition use “General Services – Education”',
                'LOAN_PAYMENTS'
            ),
            (
                'OTHER_PAYMENT',
                'Other miscellaneous debt payments',
                'LOAN_PAYMENTS'
            ),
            /* =========================  BANK_FEES  ========================== */
            (
                'ATM_FEES',
                'Fees incurred for out-of-network ATMs',
                'BANK_FEES'
            ),
            (
                'FOREIGN_TRANSACTION_FEES',
                'Fees incurred on non-domestic transactions',
                'BANK_FEES'
            ),
            (
                'INSUFFICIENT_FUNDS',
                'Fees relating to insufficient funds',
                'BANK_FEES'
            ),
            (
                'INTEREST_CHARGE',
                'Fees incurred for interest on purchases or cash advances',
                'BANK_FEES'
            ),
            (
                'OVERDRAFT_FEES',
                'Fees incurred when an account is in overdraft',
                'BANK_FEES'
            ),
            (
                'OTHER_BANK_FEES',
                'Other miscellaneous bank fees',
                'BANK_FEES'
            ),
            /* =======================  ENTERTAINMENT  ======================== */
            (
                'CASINOS_AND_GAMBLING',
                'Gambling, casinos, and sports betting',
                'ENTERTAINMENT'
            ),
            (
                'MUSIC_AND_AUDIO',
                'Digital and in-person music purchases, including music-streaming services',
                'ENTERTAINMENT'
            ),
            (
                'SPORTING_EVENTS_AMUSEMENT_PARKS_AND_MUSEUMS',
                'Purchases made at sporting events, music venues, concerts, museums, and amusement parks',
                'ENTERTAINMENT'
            ),
            (
                'TV_AND_MOVIES',
                'In-home movie-streaming services and movie theaters',
                'ENTERTAINMENT'
            ),
            (
                'VIDEO_GAMES',
                'Digital and in-person video-game purchases',
                'ENTERTAINMENT'
            ),
            (
                'OTHER_ENTERTAINMENT',
                'Other miscellaneous entertainment purchases, including night life and adult entertainment',
                'ENTERTAINMENT'
            ),
            /* =====================  FOOD_AND_DRINK  ========================= */
            (
                'BEER_WINE_AND_LIQUOR',
                'Beer, wine & liquor stores',
                'FOOD_AND_DRINK'
            ),
            (
                'COFFEE',
                'Purchases at coffee shops or cafes',
                'FOOD_AND_DRINK'
            ),
            (
                'FAST_FOOD',
                'Dining expenses for fast-food chains',
                'FOOD_AND_DRINK'
            ),
            (
                'GROCERIES',
                'Purchases for fresh produce and groceries, including farmers'' markets',
                'FOOD_AND_DRINK'
            ),
            (
                'RESTAURANT',
                'Dining expenses for restaurants, bars, gastropubs, and diners',
                'FOOD_AND_DRINK'
            ),
            (
                'VENDING_MACHINES',
                'Purchases made at vending-machine operators',
                'FOOD_AND_DRINK'
            ),
            (
                'OTHER_FOOD_AND_DRINK',
                'Other miscellaneous food and drink, including desserts, juice bars, and delis',
                'FOOD_AND_DRINK'
            ),
            /* ==================  GENERAL_MERCHANDISE  ======================= */
            (
                'BOOKSTORES_AND_NEWSSTANDS',
                'Books, magazines, and news',
                'GENERAL_MERCHANDISE'
            ),
            (
                'CLOTHING_AND_ACCESSORIES',
                'Apparel, shoes, and jewelry',
                'GENERAL_MERCHANDISE'
            ),
            (
                'CONVENIENCE_STORES',
                'Purchases at convenience stores',
                'GENERAL_MERCHANDISE'
            ),
            (
                'DEPARTMENT_STORES',
                'Retail stores with wide ranges of consumer goods, typically specializing in clothing and home goods',
                'GENERAL_MERCHANDISE'
            ),
            (
                'DISCOUNT_STORES',
                'Stores selling goods at a discounted price',
                'GENERAL_MERCHANDISE'
            ),
            (
                'ELECTRONICS',
                'Electronics stores and websites',
                'GENERAL_MERCHANDISE'
            ),
            (
                'GIFTS_AND_NOVELTIES',
                'Photo, gifts, cards, and floral stores',
                'GENERAL_MERCHANDISE'
            ),
            (
                'OFFICE_SUPPLIES',
                'Stores that specialize in office goods',
                'GENERAL_MERCHANDISE'
            ),
            (
                'ONLINE_MARKETPLACES',
                'Multi-purpose e-commerce platforms such as Etsy, eBay, and Amazon',
                'GENERAL_MERCHANDISE'
            ),
            (
                'PET_SUPPLIES',
                'Pet supplies and pet food',
                'GENERAL_MERCHANDISE'
            ),
            (
                'SPORTING_GOODS',
                'Sporting goods, camping gear, and outdoor equipment',
                'GENERAL_MERCHANDISE'
            ),
            (
                'SUPERSTORES',
                'Superstores such as Target and Walmart, selling both groceries and general merchandise',
                'GENERAL_MERCHANDISE'
            ),
            (
                'TOBACCO_AND_VAPE',
                'Purchases for tobacco and vaping products',
                'GENERAL_MERCHANDISE'
            ),
            (
                'OTHER_GENERAL_MERCHANDISE',
                'Other miscellaneous merchandise, including toys, hobbies, and arts and crafts',
                'GENERAL_MERCHANDISE'
            ),
            /* ====================  HOME_IMPROVEMENT  ======================== */
            (
                'FURNITURE',
                'Furniture, bedding, and home accessories',
                'HOME_IMPROVEMENT'
            ),
            (
                'HARDWARE',
                'Building materials, hardware stores, paint, and wallpaper',
                'HOME_IMPROVEMENT'
            ),
            (
                'REPAIR_AND_MAINTENANCE',
                'Plumbing, lighting, gardening, and roofing',
                'HOME_IMPROVEMENT'
            ),
            (
                'SECURITY',
                'Home-security-system purchases',
                'HOME_IMPROVEMENT'
            ),
            (
                'OTHER_HOME_IMPROVEMENT',
                'Other miscellaneous home purchases, including pool installation and pest control',
                'HOME_IMPROVEMENT'
            ),
            /* =========================  MEDICAL  ============================ */
            (
                'DENTAL_CARE',
                'Dentists and general dental care',
                'MEDICAL'
            ),
            (
                'EYE_CARE',
                'Optometrists, contacts, and glasses stores',
                'MEDICAL'
            ),
            (
                'NURSING_CARE',
                'Nursing care and facilities',
                'MEDICAL'
            ),
            (
                'PHARMACIES_AND_SUPPLEMENTS',
                'Pharmacies and nutrition shops',
                'MEDICAL'
            ),
            (
                'PRIMARY_CARE',
                'Doctors and physicians',
                'MEDICAL'
            ),
            (
                'VETERINARY_SERVICES',
                'Prevention and care procedures for animals',
                'MEDICAL'
            ),
            (
                'OTHER_MEDICAL',
                'Other miscellaneous medical, including blood work, hospitals, and ambulances',
                'MEDICAL'
            ),
            /* =======================  PERSONAL_CARE  ======================== */
            (
                'GYMS_AND_FITNESS_CENTERS',
                'Gyms, fitness centers, and workout classes',
                'PERSONAL_CARE'
            ),
            (
                'HAIR_AND_BEAUTY',
                'Manicures, haircuts, waxing, spa/massages, and bath and beauty products',
                'PERSONAL_CARE'
            ),
            (
                'LAUNDRY_AND_DRY_CLEANING',
                'Wash and fold, and dry-cleaning expenses',
                'PERSONAL_CARE'
            ),
            (
                'OTHER_PERSONAL_CARE',
                'Other miscellaneous personal care, including mental-health apps and services',
                'PERSONAL_CARE'
            ),
            /* =====================  GENERAL_SERVICES  ======================= */
            (
                'ACCOUNTING_AND_FINANCIAL_PLANNING',
                'Financial planning, and tax and accounting services',
                'GENERAL_SERVICES'
            ),
            (
                'AUTOMOTIVE',
                'Oil changes, car washes, repairs, and towing',
                'GENERAL_SERVICES'
            ),
            (
                'CHILDCARE',
                'Babysitters and daycare',
                'GENERAL_SERVICES'
            ),
            (
                'CONSULTING_AND_LEGAL',
                'Consulting and legal services',
                'GENERAL_SERVICES'
            ),
            (
                'EDUCATION',
                'Elementary, high-school, professional schools, and college tuition',
                'GENERAL_SERVICES'
            ),
            (
                'INSURANCE',
                'Insurance for auto, home, and healthcare',
                'GENERAL_SERVICES'
            ),
            (
                'POSTAGE_AND_SHIPPING',
                'Mail, packaging, and shipping services',
                'GENERAL_SERVICES'
            ),
            (
                'STORAGE',
                'Storage services and facilities',
                'GENERAL_SERVICES'
            ),
            (
                'OTHER_GENERAL_SERVICES',
                'Other miscellaneous services, including advertising and cloud storage',
                'GENERAL_SERVICES'
            ),
            /* ============  GOVERNMENT_AND_NON_PROFIT  ====================== */
            (
                'DONATIONS',
                'Charitable, political, and religious donations',
                'GOVERNMENT_AND_NON_PROFIT'
            ),
            (
                'GOVERNMENT_DEPARTMENTS_AND_AGENCIES',
                'Government departments and agencies, such as driving licences and passport renewal',
                'GOVERNMENT_AND_NON_PROFIT'
            ),
            (
                'TAX_PAYMENT',
                'Tax payments, including income and property taxes',
                'GOVERNMENT_AND_NON_PROFIT'
            ),
            (
                'OTHER_GOVERNMENT_AND_NON_PROFIT',
                'Other miscellaneous government and non-profit agencies',
                'GOVERNMENT_AND_NON_PROFIT'
            ),
            /* =====================  TRANSPORTATION  ========================= */
            (
                'BIKES_AND_SCOOTERS',
                'Bike and scooter rentals',
                'TRANSPORTATION'
            ),
            (
                'GAS',
                'Purchases at a gas station',
                'TRANSPORTATION'
            ),
            (
                'PARKING',
                'Parking fees and expenses',
                'TRANSPORTATION'
            ),
            (
                'PUBLIC_TRANSIT',
                'Public transportation, including rail and train, buses, and metro',
                'TRANSPORTATION'
            ),
            (
                'TAXIS_AND_RIDE_SHARES',
                'Taxi and ride-share services',
                'TRANSPORTATION'
            ),
            ('TOLLS', 'Toll expenses', 'TRANSPORTATION'),
            (
                'OTHER_TRANSPORTATION',
                'Other miscellaneous transportation expenses',
                'TRANSPORTATION'
            ),
            /* ==========================  TRAVEL  ============================ */
            ('FLIGHTS', 'Airline expenses', 'TRAVEL'),
            (
                'LODGING',
                'Hotels, motels, and hosted accommodation such as Airbnb',
                'TRAVEL'
            ),
            (
                'RENTAL_CARS',
                'Rental cars, charter buses, and trucks',
                'TRAVEL'
            ),
            (
                'OTHER_TRAVEL',
                'Other miscellaneous travel expenses',
                'TRAVEL'
            ),
            /* ====================  RENT_AND_UTILITIES  ====================== */
            (
                'GAS_AND_ELECTRICITY',
                'Gas and electricity bills',
                'RENT_AND_UTILITIES'
            ),
            (
                'INTERNET_AND_CABLE',
                'Internet and cable bills',
                'RENT_AND_UTILITIES'
            ),
            ('RENT', 'Rent payment', 'RENT_AND_UTILITIES'),
            (
                'SEWAGE_AND_WASTE_MANAGEMENT',
                'Sewage and garbage-disposal bills',
                'RENT_AND_UTILITIES'
            ),
            (
                'TELEPHONE',
                'Cell-phone bills',
                'RENT_AND_UTILITIES'
            ),
            ('WATER', 'Water bills', 'RENT_AND_UTILITIES'),
            (
                'OTHER_UTILITIES',
                'Other miscellaneous utility bills',
                'RENT_AND_UTILITIES'
            )
    )
INSERT INTO
    detailed_categories (name, description, primary_category_id)
SELECT
    s.name,
    s.description,
    p.id
FROM
    src s
    JOIN primary_categories p ON p.name = s.primary_name
ON CONFLICT (primary_category_id, name) DO NOTHING;

-- +goose Down
DELETE FROM detailed_categories
WHERE
    primary_category_id IN (
        SELECT
            id
        FROM
            primary_categories
        WHERE
            name IN (
                'INCOME',
                'TRANSFER_IN',
                'TRANSFER_OUT',
                'LOAN_PAYMENTS',
                'BANK_FEES',
                'ENTERTAINMENT',
                'FOOD_AND_DRINK',
                'GENERAL_MERCHANDISE',
                'HOME_IMPROVEMENT',
                'MEDICAL',
                'PERSONAL_CARE',
                'GENERAL_SERVICES',
                'GOVERNMENT_AND_NON_PROFIT',
                'TRANSPORTATION',
                'TRAVEL',
                'RENT_AND_UTILITIES'
            )
    );

DELETE FROM primary_categories
WHERE
    name IN (
        'INCOME',
        'TRANSFER_IN',
        'TRANSFER_OUT',
        'LOAN_PAYMENTS',
        'BANK_FEES',
        'ENTERTAINMENT',
        'FOOD_AND_DRINK',
        'GENERAL_MERCHANDISE',
        'HOME_IMPROVEMENT',
        'MEDICAL',
        'PERSONAL_CARE',
        'GENERAL_SERVICES',
        'GOVERNMENT_AND_NON_PROFIT',
        'TRANSPORTATION',
        'TRAVEL',
        'RENT_AND_UTILITIES'
    );