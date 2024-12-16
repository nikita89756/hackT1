declare @schema NVARCHAR(255) = 'SalesLT'
declare @table NVARCHAR(255) = 'Product'
declare cur cursor  FAST_FORWARD for select COLUMN_NAME  from AdventureWorksLT.INFORMATION_SCHEMA.COLUMNS
where DATA_TYPE in ('char','varchar','nchar','nvarchar','text','ntext') and TABLE_NAME = @table and TABLE_SCHEMA = @schema
declare @C NVARCHAR(255)
declare @select NVARCHAR(255)
declare @colstring NVARCHAR(255) = 'Bike'

open cur
WHILE(1=1)
BEGIN
fetch cur into @C
if @@FETCH_STATUS<>0
begin
break
end
set @select = 'select ' + @C + ' from AdventureWorksLT.'+@schema+'.'+ @table+' where '+@C + ' like ''%' + @colstring +'%'''
EXEC(@select)
END
close cur

/*
CREATE PROCEDURE SalesLT.uspFindStringInTable
@schema sysname , @table sysname , @stringToFind nvarchar(2000)
AS
declare cur cursor LOCAL FAST_FORWARD for select COLUMN_NAME  from AdventureWorksLT.INFORMATION_SCHEMA.COLUMNS
where DATA_TYPE in ('char','varchar','nchar','nvarchar','text','ntext') and TABLE_NAME = @table and TABLE_SCHEMA = @schema
declare @C NVARCHAR(255)
declare @select NVARCHAR(255)
declare @cnt int
open cur
WHILE(1=1)
BEGIN
fetch cur into @C
if @@FETCH_STATUS<>0
begin
break
end
set @select = 'select ' + @C + ' from ' +'AdventureWorksLT.'+@schema+'.'+ @table+' where '+@C + ' like ''%' + @stringToFind +'%'''
EXEC(@select)
set @cnt = @cnt+@@ROWCOUNT
END
close cur

return @cnt
*/


declare @int int
exec @int = AdventureWorksLT.SalesLT.uspFindStringInTable 'SalesLT','Product','Bike'

declare @strint NVARCHAR(255) = 'Count ' + CAST(@int as varchar)
print @strint
/*
ALTER PROCEDURE uspFindStringInAllTables
@stringToFind NVARCHAR(2000)
AS
declare cur cursor LOCAL FAST_FORWARD for SELECT DISTINCT TABLE_NAME,TABLE_SCHEMA from AdventureWorksLT.INFORMATION_SCHEMA.COLUMNS
declare @schema NVARCHAR(2000)
declare @table NVARCHAR(2000)
OPEN cur 
while (1=1)
begin
declare @cnt int = 0
fetch cur into @table , @schema
if @@FETCH_STATUS <>0 
BEGIN
break
end
BEGIN TRY
exec @cnt = AdventureWorksLT.SalesLT.uspFindStringInTable @schema , @table , @stringToFind
END TRY
BEGIN CATCH
print ERROR_MESSAGE();
END CATCH;
if @cnt<>0
BEGIN
print 'in table '+@schema+'.'+@table+' count:'+CAST(@cnt as VARCHAR)
end
ELSE
BEGIN
print 'in table '+@schema+'.'+@table+' no such strings'
END
end
close cur
*/


exec AdventureWorksLT.dbo.uspFindStringInAllTables 'Bike'