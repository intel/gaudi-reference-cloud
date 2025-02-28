import { CSVLink } from 'react-csv'

const ExportDataToCSV = (props) => {
  const filename = props.filename ?? 'coupons.csv'

  return (
    <div className='text-right' intc-id='btn-export'>
      <CSVLink
        data={props.data}
        headers={props.csvHeaders}
        filename={filename}
        variant="link"
        className="btn btn-primary"
      >
        Export to CSV
      </CSVLink>
    </div>
  )
}

export default ExportDataToCSV
