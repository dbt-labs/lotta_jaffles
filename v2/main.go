package main

// imports
import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// constants and constant variables
const maxThreads = 2 << 6
const copies = 2000

var delim = "_"
var models = "models"
var staging = "staging"

// main function
func main() {

	// create the models/ directory
	if err := os.MkdirAll(models, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	// create the models/staging directory
	if err := os.MkdirAll(filepath.Join(models, staging), os.ModePerm); err != nil {
		log.Fatal(err)
	}

	// create a wait group to ensure all threads finish
	var wg sync.WaitGroup

	// create a channel to limit concurrent threads
	ch := make(chan struct{}, maxThreads)

	// create the copies
	for i := 0; i < copies; i++ {

		// add to the wait group
		wg.Add(1)

		// add to the channel
		ch <- struct{}{}

		go func(id int) {
			// remove from the wait group
			defer wg.Done()

			// debug
			log.Println("task id:", id)

			// create the model copies
			err := copyModels(expandId(id))
			if err != nil {
				log.Fatal(err)
			}

			// remove from the channel
			<-ch

		}(i)
	}
	// wait for all threads to finish
	wg.Wait()
}

// helper functions
func copyModels(id string) (err error) {

	// create the staging models

	// stg_customers
	stg_customers_sql := stgCustomersSql()

	filename := fmt.Sprintf("stg_customers%s%s.sql", delim, id)

	if err := os.WriteFile(filepath.Join(models, staging, filename), []byte(stg_customers_sql), os.ModePerm); err != nil {
		return err
	}

	// stg_orders
	stg_orders_sql := stgOrdersSql()

	filename = fmt.Sprintf("stg_orders%s%s.sql", delim, id)

	if err := os.WriteFile(filepath.Join(models, staging, filename), []byte(stg_orders_sql), os.ModePerm); err != nil {
		return err
	}

	// stg_payments
	stg_payments_sql := stgPaymentsSql()

	filename = fmt.Sprintf("stg_payments%s%s.sql", delim, id)

	if err := os.WriteFile(filepath.Join(models, staging, filename), []byte(stg_payments_sql), os.ModePerm); err != nil {
		return err
	}

	// create the marts

	// orders
	orders_sql := ordersSql()

	orders_sql = strings.Replace(orders_sql, "$DELIM", delim, -1)
	orders_sql = strings.Replace(orders_sql, "$ID", id, -1)

	filename = fmt.Sprintf("orders%s%s.sql", delim, id)

	if err := os.WriteFile(filepath.Join(models, filename), []byte(orders_sql), os.ModePerm); err != nil {
		return err
	}

	// customers
	customers_sql := customersSql()

	customers_sql = strings.Replace(customers_sql, "$DELIM", delim, -1)
	customers_sql = strings.Replace(customers_sql, "$ID", id, -1)

	filename = fmt.Sprintf("customers%s%s.sql", delim, id)

	if err := os.WriteFile(filepath.Join(models, filename), []byte(customers_sql), os.ModePerm); err != nil {
		return err
	}

	return
}

func stgCustomersSql() (stg_customers_sql string) {

	stg_customers_sql = `
with source as (

    {#-
    Normally we would select from the table here, but we are using seeds to load
    our data in this project
    #}
    select * from {{ ref('raw_customers') }}

),

renamed as (

    select
        id as customer_id,
        first_name,
        last_name

    from source

)

select * from renamed
`

	return
}

func stgOrdersSql() (stg_orders_sql string) {

	stg_orders_sql = `
with source as (

    {#-
    Normally we would select from the table here, but we are using seeds to load
    our data in this project
    #}
    select * from {{ ref('raw_orders') }}

),

renamed as (

    select
        id as order_id,
        user_id as customer_id,
        order_date,
        status

    from source

)

select * from renamed
`

	return
}

func stgPaymentsSql() (stg_payments_sql string) {

	stg_payments_sql = `
with source as (
    
    {#-
    Normally we would select from the table here, but we are using seeds to load
    our data in this project
    #}
    select * from {{ ref('raw_payments') }}

),

renamed as (

    select
        id as payment_id,
        order_id,
        payment_method,

        -- "amount" is currently stored in cents, so we convert it to dollars
        amount / 100 as amount

    from source

)

select * from renamed
`

	return
}

func ordersSql() (orders_sql string) {
	orders_sql = `
{% set payment_methods = ['credit_card', 'coupon', 'bank_transfer', 'gift_card'] %}

with orders as (

    select * from {{ ref('stg_orders$DELIM$ID') }}

),

payments as (

    select * from {{ ref('stg_payments$DELIM$ID') }}

),

order_payments as (

    select
        order_id,

        {% for payment_method in payment_methods -%}
        sum(case when payment_method = '{{ payment_method }}' then amount else 0 end) as {{ payment_method }}_amount,
        {% endfor -%}

        sum(amount) as total_amount

    from payments

    group by order_id

),

final as (

    select
        orders.order_id,
        orders.customer_id,
        orders.order_date,
        orders.status,

        {% for payment_method in payment_methods -%}

        order_payments.{{ payment_method }}_amount,

        {% endfor -%}

        order_payments.total_amount as amount

    from orders


    left join order_payments
        on orders.order_id = order_payments.order_id

)

select * from final
`

	return
}

func customersSql() (customers_sql string) {

	customers_sql = `
with orders as (

	select * from {{ ref('orders$DELIM$ID') }}

),

customers as (

	select * from {{ ref('stg_customers$DELIM$ID') }}

),

final as (

	select
		customers.customer_id,
		customers.first_name,
		customers.last_name,

		count(orders.order_id) as orders_count,

		sum(orders.amount) as total_amount

	from customers

	left join orders
		on customers.customer_id = orders.customer_id

	group by 1, 2, 3

)

select * from final
`

	return
}

func expandId(id int) (expanded string) {

	// use number of digits in copies to determine the number of zeros to prepend
	expanded = fmt.Sprintf("%0*d", len(fmt.Sprintf("%d", copies-1)), id)

	return
}
