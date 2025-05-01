-- +goose Up
-- Insert unique primary categories.
INSERT
    OR IGNORE INTO primary_categories (name)
VALUES ('INCOME'),
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
    ('RENT_AND_UTILITIES');
-- Insert detailed categories.
-- For each row, we remove the primary prefix and underscore from the DETAILED value.
INSERT INTO detailed_categories (name, primary_category_id, description)
VALUES (
        'DIVIDENDS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'INCOME'
        ),
        'Dividends from investment accounts'
    ),
    (
        'INTEREST_EARNED',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'INCOME'
        ),
        'Income from interest on savings accounts'
    ),
    (
        'RETIREMENT_PENSION',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'INCOME'
        ),
        'Income from pension payments'
    ),
    (
        'TAX_REFUND',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'INCOME'
        ),
        'Income from tax refunds'
    ),
    (
        'UNEMPLOYMENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'INCOME'
        ),
        'Income from unemployment benefits, including unemployment insurance and healthcare'
    ),
    (
        'WAGES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'INCOME'
        ),
        'Income from salaries, gig-economy work, and tips earned'
    ),
    (
        'OTHER_INCOME',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'INCOME'
        ),
        'Other miscellaneous income, including alimony, social security, child support, and rental'
    ),
    (
        'CASH_ADVANCES_AND_LOANS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSFER_IN'
        ),
        'Loans and cash advances deposited into a bank account'
    ),
    (
        'DEPOSIT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSFER_IN'
        ),
        'Cash, checks, and ATM deposits into a bank account'
    ),
    (
        'INVESTMENT_AND_RETIREMENT_FUNDS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSFER_IN'
        ),
        'Inbound transfers to an investment or retirement account'
    ),
    (
        'SAVINGS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSFER_IN'
        ),
        'Inbound transfers to a savings account'
    ),
    (
        'ACCOUNT_TRANSFER',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSFER_IN'
        ),
        'General inbound transfers from another account'
    ),
    (
        'OTHER_TRANSFER_IN',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSFER_IN'
        ),
        'Other miscellaneous inbound transactions'
    ),
    (
        'INVESTMENT_AND_RETIREMENT_FUNDS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSFER_OUT'
        ),
        'Transfers to an investment or retirement account, including investment apps such as Acorns, Betterment'
    ),
    (
        'SAVINGS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSFER_OUT'
        ),
        'Outbound transfers to savings accounts'
    ),
    (
        'WITHDRAWAL',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSFER_OUT'
        ),
        'Withdrawals from a bank account'
    ),
    (
        'ACCOUNT_TRANSFER',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSFER_OUT'
        ),
        'General outbound transfers to another account'
    ),
    (
        'OTHER_TRANSFER_OUT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSFER_OUT'
        ),
        'Other miscellaneous outbound transactions'
    ),
    (
        'CAR_PAYMENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'LOAN_PAYMENTS'
        ),
        'Car loans and leases'
    ),
    (
        'CREDIT_CARD_PAYMENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'LOAN_PAYMENTS'
        ),
        'Payments to a credit card. These are positive amounts for credit card subtypes and negative for depository subtypes'
    ),
    (
        'PERSONAL_LOAN_PAYMENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'LOAN_PAYMENTS'
        ),
        'Personal loans, including cash advances and buy now pay later repayments'
    ),
    (
        'MORTGAGE_PAYMENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'LOAN_PAYMENTS'
        ),
        'Payments on mortgages'
    ),
    (
        'STUDENT_LOAN_PAYMENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'LOAN_PAYMENTS'
        ),
        'Payments on student loans. For college tuition, refer to ''General Services - Education'''
    ),
    (
        'OTHER_PAYMENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'LOAN_PAYMENTS'
        ),
        'Other miscellaneous debt payments'
    ),
    (
        'ATM_FEES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'BANK_FEES'
        ),
        'Fees incurred for out-of-network ATMs'
    ),
    (
        'FOREIGN_TRANSACTION_FEES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'BANK_FEES'
        ),
        'Fees incurred on non-domestic transactions'
    ),
    (
        'INSUFFICIENT_FUNDS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'BANK_FEES'
        ),
        'Fees relating to insufficient funds'
    ),
    (
        'INTEREST_CHARGE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'BANK_FEES'
        ),
        'Fees incurred for interest on purchases, including not-paid-in-full or interest on cash advances'
    ),
    (
        'OVERDRAFT_FEES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'BANK_FEES'
        ),
        'Fees incurred when an account is in overdraft'
    ),
    (
        'OTHER_BANK_FEES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'BANK_FEES'
        ),
        'Other miscellaneous bank fees'
    ),
    (
        'CASINOS_AND_GAMBLING',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'ENTERTAINMENT'
        ),
        'Gambling, casinos, and sports betting'
    ),
    (
        'MUSIC_AND_AUDIO',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'ENTERTAINMENT'
        ),
        'Digital and in-person music purchases, including music streaming services'
    ),
    (
        'SPORTING_EVENTS_AMUSEMENT_PARKS_AND_MUSEUMS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'ENTERTAINMENT'
        ),
        'Purchases made at sporting events, music venues, concerts, museums, and amusement parks'
    ),
    (
        'TV_AND_MOVIES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'ENTERTAINMENT'
        ),
        'In home movie streaming services and movie theaters'
    ),
    (
        'VIDEO_GAMES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'ENTERTAINMENT'
        ),
        'Digital and in-person video game purchases'
    ),
    (
        'OTHER_ENTERTAINMENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'ENTERTAINMENT'
        ),
        'Other miscellaneous entertainment purchases, including night life and adult entertainment'
    ),
    (
        'BEER_WINE_AND_LIQUOR',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'FOOD_AND_DRINK'
        ),
        'Beer, Wine & Liquor Stores'
    ),
    (
        'COFFEE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'FOOD_AND_DRINK'
        ),
        'Purchases at coffee shops or cafes'
    ),
    (
        'FAST_FOOD',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'FOOD_AND_DRINK'
        ),
        'Dining expenses for fast food chains'
    ),
    (
        'GROCERIES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'FOOD_AND_DRINK'
        ),
        'Purchases for fresh produce and groceries, including farmers'' markets'
    ),
    (
        'RESTAURANT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'FOOD_AND_DRINK'
        ),
        'Dining expenses for restaurants, bars, gastropubs, and diners'
    ),
    (
        'VENDING_MACHINES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'FOOD_AND_DRINK'
        ),
        'Purchases made at vending machine operators'
    ),
    (
        'OTHER_FOOD_AND_DRINK',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'FOOD_AND_DRINK'
        ),
        'Other miscellaneous food and drink, including desserts, juice bars, and delis'
    ),
    (
        'BOOKSTORES_AND_NEWSSTANDS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Books, magazines, and news'
    ),
    (
        'CLOTHING_AND_ACCESSORIES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Apparel, shoes, and jewelry'
    ),
    (
        'CONVENIENCE_STORES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Purchases at convenience stores'
    ),
    (
        'DEPARTMENT_STORES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Retail stores with wide ranges of consumer goods, typically specializing in clothing and home goods'
    ),
    (
        'DISCOUNT_STORES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Stores selling goods at a discounted price'
    ),
    (
        'ELECTRONICS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Electronics stores and websites'
    ),
    (
        'GIFTS_AND_NOVELTIES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Photo, gifts, cards, and floral stores'
    ),
    (
        'OFFICE_SUPPLIES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Stores that specialize in office goods'
    ),
    (
        'ONLINE_MARKETPLACES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Multi-purpose e-commerce platforms such as Etsy, Ebay and Amazon'
    ),
    (
        'PET_SUPPLIES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Pet supplies and pet food'
    ),
    (
        'SPORTING_GOODS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Sporting goods, camping gear, and outdoor equipment'
    ),
    (
        'SUPERSTORES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Superstores such as Target and Walmart, selling both groceries and general merchandise'
    ),
    (
        'TOBACCO_AND_VAPE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Purchases for tobacco and vaping products'
    ),
    (
        'OTHER_GENERAL_MERCHANDISE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_MERCHANDISE'
        ),
        'Other miscellaneous merchandise, including toys, hobbies, and arts and crafts'
    ),
    (
        'FURNITURE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'HOME_IMPROVEMENT'
        ),
        'Furniture, bedding, and home accessories'
    ),
    (
        'HARDWARE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'HOME_IMPROVEMENT'
        ),
        'Building materials, hardware stores, paint, and wallpaper'
    ),
    (
        'REPAIR_AND_MAINTENANCE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'HOME_IMPROVEMENT'
        ),
        'Plumbing, lighting, gardening, and roofing'
    ),
    (
        'SECURITY',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'HOME_IMPROVEMENT'
        ),
        'Home security system purchases'
    ),
    (
        'OTHER_HOME_IMPROVEMENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'HOME_IMPROVEMENT'
        ),
        'Other miscellaneous home purchases, including pool installation and pest control'
    ),
    (
        'DENTAL_CARE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'MEDICAL'
        ),
        'Dentists and general dental care'
    ),
    (
        'EYE_CARE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'MEDICAL'
        ),
        'Optometrists, contacts, and glasses stores'
    ),
    (
        'NURSING_CARE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'MEDICAL'
        ),
        'Nursing care and facilities'
    ),
    (
        'PHARMACIES_AND_SUPPLEMENTS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'MEDICAL'
        ),
        'Pharmacies and nutrition shops'
    ),
    (
        'PRIMARY_CARE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'MEDICAL'
        ),
        'Doctors and physicians'
    ),
    (
        'VETERINARY_SERVICES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'MEDICAL'
        ),
        'Prevention and care procedures for animals'
    ),
    (
        'OTHER_MEDICAL',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'MEDICAL'
        ),
        'Other miscellaneous medical, including blood work, hospitals, and ambulances'
    ),
    (
        'GYMS_AND_FITNESS_CENTERS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'PERSONAL_CARE'
        ),
        'Gyms, fitness centers, and workout classes'
    ),
    (
        'HAIR_AND_BEAUTY',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'PERSONAL_CARE'
        ),
        'Manicures, haircuts, waxing, spa/massages, and bath and beauty products'
    ),
    (
        'LAUNDRY_AND_DRY_CLEANING',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'PERSONAL_CARE'
        ),
        'Wash and fold, and dry cleaning expenses'
    ),
    (
        'OTHER_PERSONAL_CARE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'PERSONAL_CARE'
        ),
        'Other miscellaneous personal care, including mental health apps and services'
    ),
    (
        'ACCOUNTING_AND_FINANCIAL_PLANNING',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_SERVICES'
        ),
        'Financial planning, and tax and accounting services'
    ),
    (
        'AUTOMOTIVE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_SERVICES'
        ),
        'Oil changes, car washes, repairs, and towing'
    ),
    (
        'CHILDCARE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_SERVICES'
        ),
        'Babysitters and daycare'
    ),
    (
        'CONSULTING_AND_LEGAL',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_SERVICES'
        ),
        'Consulting and legal services'
    ),
    (
        'EDUCATION',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_SERVICES'
        ),
        'Elementary, high school, professional schools, and college tuition'
    ),
    (
        'INSURANCE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_SERVICES'
        ),
        'Insurance for auto, home, and healthcare'
    ),
    (
        'POSTAGE_AND_SHIPPING',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_SERVICES'
        ),
        'Mail, packaging, and shipping services'
    ),
    (
        'STORAGE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_SERVICES'
        ),
        'Storage services and facilities'
    ),
    (
        'OTHER_GENERAL_SERVICES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GENERAL_SERVICES'
        ),
        'Other miscellaneous services, including advertising and cloud storage'
    ),
    (
        'DONATIONS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GOVERNMENT_AND_NON_PROFIT'
        ),
        'Charitable, political, and religious donations'
    ),
    (
        'GOVERNMENT_DEPARTMENTS_AND_AGENCIES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GOVERNMENT_AND_NON_PROFIT'
        ),
        'Government departments and agencies, such as driving licences, and passport renewal'
    ),
    (
        'TAX_PAYMENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GOVERNMENT_AND_NON_PROFIT'
        ),
        'Tax payments, including income and property taxes'
    ),
    (
        'OTHER_GOVERNMENT_AND_NON_PROFIT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'GOVERNMENT_AND_NON_PROFIT'
        ),
        'Other miscellaneous government and non-profit agencies'
    ),
    (
        'BIKES_AND_SCOOTERS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSPORTATION'
        ),
        'Bike and scooter rentals'
    ),
    (
        'GAS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSPORTATION'
        ),
        'Purchases at a gas station'
    ),
    (
        'PARKING',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSPORTATION'
        ),
        'Parking fees and expenses'
    ),
    (
        'PUBLIC_TRANSIT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSPORTATION'
        ),
        'Public transportation, including rail and train, buses, and metro'
    ),
    (
        'TAXIS_AND_RIDE_SHARES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSPORTATION'
        ),
        'Taxi and ride share services'
    ),
    (
        'TOLLS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSPORTATION'
        ),
        'Toll expenses'
    ),
    (
        'OTHER_TRANSPORTATION',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRANSPORTATION'
        ),
        'Other miscellaneous transportation expenses'
    ),
    (
        'FLIGHTS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRAVEL'
        ),
        'Airline expenses'
    ),
    (
        'LODGING',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRAVEL'
        ),
        'Hotels, motels, and hosted accommodation such as Airbnb'
    ),
    (
        'RENTAL_CARS',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRAVEL'
        ),
        'Rental cars, charter buses, and trucks'
    ),
    (
        'OTHER_TRAVEL',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'TRAVEL'
        ),
        'Other miscellaneous travel expenses'
    ),
    (
        'GAS_AND_ELECTRICITY',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'RENT_AND_UTILITIES'
        ),
        'Gas and electricity bills'
    ),
    (
        'INTERNET_AND_CABLE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'RENT_AND_UTILITIES'
        ),
        'Internet and cable bills'
    ),
    (
        'RENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'RENT_AND_UTILITIES'
        ),
        'Rent payment'
    ),
    (
        'SEWAGE_AND_WASTE_MANAGEMENT',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'RENT_AND_UTILITIES'
        ),
        'Sewage and garbage disposal bills'
    ),
    (
        'TELEPHONE',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'RENT_AND_UTILITIES'
        ),
        'Cell phone bills'
    ),
    (
        'WATER',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'RENT_AND_UTILITIES'
        ),
        'Water bills'
    ),
    (
        'OTHER_UTILITIES',
        (
            SELECT id
            FROM primary_categories
            WHERE name = 'RENT_AND_UTILITIES'
        ),
        'Other miscellaneous utility bills'
    );
-- +goose Down
-- Remove detailed categories first, then primary categories.
DELETE FROM detailed_categories
WHERE primary_category_id IN (
        SELECT id
        FROM primary_categories
        WHERE name IN (
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
WHERE name IN (
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