var Fabric_Client = require('fabric-client');
var path = require('path');
var util = require('util');
var os = require('os');

var fabric_client = new Fabric_Client();

var channel = fabric_client.newChannel('mychannel');
var peer = fabric_client.newPeer('grpc://localhost:7051');
channel.addPeer(peer);
var order = fabric_client.newOrderer('grpc://localhost:7050')
channel.addOrderer(order);

var store_path = path.join(__dirname, 'hfc-key-store');
console.log('Store path:'+store_path);
var tx_id = null;

const express = require('express')
const app = express()
const port = 3000
const bodyParser = require('body-parser');
const cors = require('cors')

app.use(cors());
app.use(bodyParser.urlencoded({ extended: true }));
app.get('/test', (req, res) => res.send('Hello World!'))


app.listen(port, () => console.log(`Example app listening on port ${port}!`))

app.all('/invoice', function(req, res){    

Fabric_Client.newDefaultKeyValueStore({ path: store_path
}).then((state_store) => {

fabric_client.setStateStore(state_store);
var crypto_suite = Fabric_Client.newCryptoSuite();

var crypto_store = Fabric_Client.newCryptoKeyStore({path: store_path});
crypto_suite.setCryptoKeyStore(crypto_store);
fabric_client.setCryptoSuite(crypto_suite);

//Para sa getUserContext
var username = req.body.username;
console.log(username);
return fabric_client.getUserContext(username, true);
}).then((user_from_store) => {
  if (user_from_store && user_from_store.isEnrolled()) {
    var username = req.body.username;
    console.log('Successfully loaded ' + username + ' from persistence');
    member_user = user_from_store;
  } else {
    var username = req.body.username;
  throw new Error('Failed to get ' + username + '. Run registerUser.js');
}

tx_id = fabric_client.newTransactionID();
console.log("Assigning transaction_id: ", tx_id._transaction_id);

var request = {
  chaincodeId: 'invoice',
  chainId: 'mychannel',
  txId: tx_id
};

var newInvoice = [];
var invoiceId = req.body.invoiceId;
var invoiceNumber = req.body.invoiceNumber;
var billedTo = req.body.billedTo;
var invoiceDate = req.body.invoiceDate;
var invoiceAmount = req.body.invoiceAmount;
var itemDescription = req.body.itemDescription;
var gr = req.body.gr;
var isPaid = req.body.isPaid;
var paidAmount = req.body.paidAmount;
var isRepaid = req.body.isRepaid;
var repaymentAmount = req.body.repaymentAmount;

newInvoice.push(invoiceId);
if (req.method == "POST")
{
  var username = req.body.username;

  if (username != "IBM"){
    res.json(username + " is not allowed to do this transaction.");
    throw new Error(username + " is not allowed to do this transaction.");
  }

  request.fcn='createInvoice';
  newInvoice.push(invoiceNumber);
  newInvoice.push(billedTo);
  newInvoice.push(invoiceDate);
  newInvoice.push(invoiceAmount); 
  newInvoice.push(itemDescription);
}

else if(req.method == "PUT")
{
  if(gr)    
  {
    var username = req.body.username;

    if (username != "YKK"){
      res.json(username + " is not allowed to do this transaction.");
      throw new Error(username + " is not allowed to do this transaction.");
    } else {
      request.fcn = 'isGoodsReceived',
      newInvoice.push(gr);
    }
    
  }
  
  else if(paidAmount)
  {

    var username = req.body.username;

    if (username != "Union"){

      res.json(username + " is not allowed to do this transaction.");
      throw new Error(username + " is not allowed to do this transaction.");
    } else {
      request.fcn = 'isPaidToSupplier',
      newInvoice.push(paidAmount);
    }

    
  }

  else if(repaymentAmount)
  {

    var username = req.body.username;

    if (username != "YKK"){
      
      res.json(username + " is not allowed to do this transaction.");
      throw new Error(username + " is not allowed to do this transaction.");
    } else {
      request.fcn = 'isRepaidToBank',
      newInvoice.push(repaymentAmount);
    }
    
  }
}


request.args=newInvoice;
console.log(request);

res.json ({
  Function: request.fcn,
  Inputs: request.args,
  Result: "Success"
});

return channel.sendTransactionProposal(request);
}).then((results) => {
var proposalResponses = results[0];
var proposal = results[1];
let isProposalGood = false;
if (proposalResponses && proposalResponses[0].response &&
proposalResponses[0].response.status === 200) {
isProposalGood = true;
console.log('Transaction proposal was good');
} else {
console.error('Transaction proposal was bad');
}
if (isProposalGood) {
console.log(util.format(
'Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s"',
proposalResponses[0].response.status, proposalResponses[0].response.message));

var request = {
proposalResponses: proposalResponses,
proposal: proposal
};

var transaction_id_string = tx_id.getTransactionID(); //Get the transaction ID string to be used by the event processing
var promises = [];

var sendPromise = channel.sendTransaction(request);
promises.push(sendPromise); 

let event_hub = channel.newChannelEventHub(peer);

let txPromise = new Promise((resolve, reject) => {
let handle = setTimeout(() => {
event_hub.unregisterTxEvent(transaction_id_string);
event_hub.disconnect();
resolve({event_status : 'TIMEOUT'});
}, 3000);
event_hub.registerTxEvent(transaction_id_string, (tx, code) => {
clearTimeout(handle);


var return_status = {event_status : code, tx_id : transaction_id_string};
if (code !== 'VALID') {
console.error('The transaction was invalid, code = ' + code);
resolve(return_status);
} else {
console.log('The transaction has been committed on peer ' + event_hub.getPeerAddr());
resolve(return_status);
}
}, (err) => {
reject(new Error('There was a problem with the eventhub ::'+err));
},
{disconnect: true} //disconnect when complete
);
event_hub.connect();

});
promises.push(txPromise);

return Promise.all(promises);
} else {
res.json({"lol":"butts"});
console.error('Failed to send Proposal or receive valid response. Response null or status is not 200. exiting...');
throw new Error('Failed to send Proposal or receive valid response. Response null or status is not 200. exiting...');
}
}).then((results) => {
console.log('Send transaction promise and event listener promise have completed');
// check the results in the order the promises were added to the promise all list
if (results && results[0] && results[0].status === 'SUCCESS') {
console.log('Successfully sent transaction to the orderer.');
} else {
console.error('Failed to order the transaction. Error code: ' + results[0].status);
}

if(results && results[1] && results[1].event_status === 'VALID') {
console.log('Successfully committed the change to the ledger by the peer');
                res.json({'result': 'success'});
} else {
console.log('Transaction failed to be committed to the ledger due to ::'+results[1].event_status);
}
}).catch((err) => {
console.error('Failed to invoke successfully :: ' + err);
});


})

app.get('/', function (req, res) {

Fabric_Client.newDefaultKeyValueStore({ path: store_path
}).then((state_store) => {
// assign the store to the fabric client
fabric_client.setStateStore(state_store);
var crypto_suite = Fabric_Client.newCryptoSuite();
// use the same location for the state store (where the users' certificate are kept)
// and the crypto store (where the users' keys are kept)
var crypto_store = Fabric_Client.newCryptoKeyStore({path: store_path});
crypto_suite.setCryptoKeyStore(crypto_store);
fabric_client.setCryptoSuite(crypto_suite);

var username = req.body.username;
console.log(username);

return fabric_client.getUserContext(username, true);
}).then((user_from_store) => {
  if (user_from_store && user_from_store.isEnrolled()) {
    var username = req.body.username;
    console.log('Successfully loaded ' + username + ' from persistence');
    member_user = user_from_store;
  } else {
    var username = req.body.username;
  throw new Error('Failed to get ' + username + '. Run registerUser.js');
}

const request = {
  chaincodeId: 'invoice',
  fcn: 'displayAllInvoices',
  args: ['']
};

var ar = [];
var attr = req.query.attr;
var invoiceId = req.query.invoiceId;

if (attr){
  ar.push(attr);
  request.fcn = 'getUser';
  request.args = ar;
} else if (invoiceId){
  ar.push(invoiceId);
  request.fcn = 'getAuditHistoryForInvoice';
  request.args = ar;

}

// send the query proposal to the peer
return channel.queryByChaincode(request);
}).then((query_responses) => {
console.log("Query has completed, checking results");
// query_responses could have more than one  results if there multiple peers were used as targets
if (query_responses && query_responses.length == 1) {
if (query_responses[0] instanceof Error) {
console.error("error from query = ", query_responses[0]);
} else {
console.log("Response is ", query_responses[0].toString());
                        res.send(query_responses[0].toString());
}
} else {
console.log("No payloads were returned from query");
}
}).catch((err) => {
console.error('Failed to query successfully :: ' + err);
});

})

//const block = channel.queryInfo(peer,false);
//console.log("height:"+block.height);

  function unicodeToChar(text) {
    return text.replace(/\\u[\dA-F]{4}/gi, 
           function (match) {
                return String.fromCharCode(parseInt(match.replace(/\\u/g, ''), 16));
           });
 }
  


  
  